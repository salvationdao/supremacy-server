package voice_chat

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"net/url"
	"server"
	"server/db"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"strings"
	"time"

	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/ws"
	"github.com/sasha-s/go-deadlock"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"golang.org/x/exp/slices"
)

type VoiceChannel struct {
	Boston      []*boiler.VoiceStream
	Zaibatsu    []*boiler.VoiceStream
	RedMountain []*boiler.VoiceStream

	deadlock.RWMutex
}

type SignedPolicyURL struct {
	ListenURL string
	SendURL   string
	ExpiredAt time.Time
}

var VoiceChatSecretKey string

func GetSignedPolicyURL(ownerID string) (*SignedPolicyURL, error) {
	baseURL := fmt.Sprintf("%s/%s", db.GetStrWithDefault(db.KeyOvenmediaStreamURL, "wss://stream.supremacygame.io:3334/app"), ownerID)
	urlExpiryTime := db.GetIntWithDefault(db.KeyVoiceExpiryTimeHours, 2000)
	signedPolicyURL := &SignedPolicyURL{}

	expiryTime := time.Now().Add(time.Hour * time.Duration(urlExpiryTime))
	signedPolicyURL.ExpiredAt = expiryTime

	sendURL, err := generateSignedURL(baseURL, expiryTime, true)
	if err != nil {
		gamelog.L.Error().Msg("failed to generate signed url for sending")
		return nil, terror.Error(err, "failed to generate signed url for sending")
	}

	listenURL, err := generateSignedURL(baseURL, expiryTime, false)
	if err != nil {
		gamelog.L.Error().Msg("failed to generate signed url for listening")
		return nil, terror.Error(err, "failed to generate signed url for listening")
	}

	signedPolicyURL.SendURL = sendURL
	signedPolicyURL.ListenURL = listenURL

	return signedPolicyURL, nil
}

func (vc *VoiceChannel) UpdateAllVoiceChannel(warMachineIDs []string, arenaID string) error {
	vc.Lock()
	defer vc.Unlock()

	ci, err := boiler.CollectionItems(
		boiler.CollectionItemWhere.ItemID.IN(warMachineIDs),
		boiler.CollectionItemWhere.ItemType.EQ(boiler.ItemTypeMech),
		qm.Load(boiler.CollectionItemRels.Owner),
	).All(gamedb.StdConn)
	if err != nil {
		return err
	}

	_, err = boiler.VoiceStreams(
		boiler.VoiceStreamWhere.ArenaID.EQ(arenaID),
		boiler.VoiceStreamWhere.IsActive.EQ(true),
		boiler.VoiceStreamWhere.SenderType.EQ(boiler.VoiceSenderTypeMECH_OWNER),
	).UpdateAll(gamedb.StdConn, boiler.M{
		boiler.VoiceStreamColumns.IsActive: false,
	})
	if err != nil {
		return terror.Error(err, "Failed to update current active")
	}

	var zaiChannel []*boiler.VoiceStream
	var bostonChannel []*boiler.VoiceStream
	var rmChannel []*boiler.VoiceStream

	checkList := []string{}

	for _, machineStream := range ci {
		if slices.Index(checkList, machineStream.OwnerID) != -1 || machineStream.R.Owner.IsAi == true {
			continue
		}

		checkList = append(checkList, machineStream.OwnerID)

		policyURL, err := GetSignedPolicyURL(machineStream.OwnerID)
		if err != nil {
			gamelog.L.Error().Str("owner_id", machineStream.OwnerID).Err(err).Msg("Failed to get signed policy url")
			continue
		}

		voiceStream := &boiler.VoiceStream{
			ArenaID:         arenaID,
			OwnerID:         machineStream.OwnerID,
			FactionID:       machineStream.R.Owner.FactionID.String,
			IsActive:        true,
			SenderType:      boiler.VoiceSenderTypeMECH_OWNER,
			SendStreamURL:   policyURL.SendURL,
			ListenStreamURL: policyURL.ListenURL,
			SessionExpireAt: policyURL.ExpiredAt,
		}

		err = voiceStream.Insert(gamedb.StdConn, boil.Infer())
		if err != nil {
			gamelog.L.Error().Str("owner_id", machineStream.OwnerID).Err(err).Msg("Failed to insert voice stream")
			continue
		}

		switch machineStream.R.Owner.FactionID.String {
		case server.ZaibatsuFactionID:
			zaiChannel = append(zaiChannel, voiceStream)
		case server.RedMountainFactionID:
			rmChannel = append(rmChannel, voiceStream)
		case server.BostonCyberneticsFactionID:
			bostonChannel = append(bostonChannel, voiceStream)
		}
	}

	ps, err := boiler.Players(
		boiler.PlayerWhere.ID.IN(ws.TrackedIdents()),
		boiler.PlayerWhere.FactionID.IsNotNull(),
		boiler.PlayerWhere.IsAi.EQ(false),
	).All(gamedb.StdConn)
	if err != nil {
		return err
	}

	for _, p := range ps {
		vcs := []*server.VoiceStreamResp{}
		switch p.FactionID.String {
		case server.ZaibatsuFactionID:
			vcs, err = db.GetActiveVoiceChat(p.ID, p.FactionID.String, arenaID)
			if err != nil {
				return terror.Error(err, "Failed to get active voice chat")

			}
		case server.RedMountainFactionID:
			vcs, err = db.GetActiveVoiceChat(p.ID, p.FactionID.String, arenaID)
			if err != nil {
				return terror.Error(err, "Failed to get active voice chat")

			}
		case server.BostonCyberneticsFactionID:
			vcs, err = db.GetActiveVoiceChat(p.ID, p.FactionID.String, arenaID)
			if err != nil {
				return terror.Error(err, "Failed to get active voice chat")

			}
		}

		ws.PublishMessage(fmt.Sprintf("/secure/user/%s/arena/%s", p.ID, arenaID), server.HubKeyVoiceStreams, vcs)
	}

	return nil
}

func UpdateFactionVoiceChannel(factionID, arenaID string) error {
	ps, err := boiler.Players(
		qm.Select(boiler.PlayerColumns.ID, boiler.PlayerColumns.FactionID),
		boiler.PlayerWhere.ID.IN(ws.TrackedIdents()),
		boiler.PlayerWhere.FactionID.IsNotNull(),
	).All(gamedb.StdConn)
	if err != nil {
		return err
	}

	for _, p := range ps {
		vcs, err := db.GetActiveVoiceChat(p.ID, p.FactionID.String, arenaID)
		if err != nil {
			return terror.Error(err, "Failed to get active voice chat")

		}

		ws.PublishMessage(fmt.Sprintf("/secure/user/%s/arena/%s", p.ID, arenaID), server.HubKeyVoiceStreams, vcs)
	}

	return nil
}

func generateSignedURL(baseURL string, expiryTime time.Time, send bool) (string, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return "", terror.Error(err, "Failed to parse base url")
	}
	policy := fmt.Sprintf("{\"url_expire\":%d}", expiryTime.UnixMilli())
	encodedPolicy := removeEncodePadding(base64.StdEncoding.EncodeToString([]byte(policy)))
	query := u.Query()
	if send {
		query.Add("direction", "send")
	}

	query.Add("policy", encodedPolicy)
	u.RawQuery = query.Encode()
	// remove percent encode
	decoded, err := url.QueryUnescape(query.Encode())
	if err != nil {
		gamelog.L.Error().Msg("Failed to decode url")
		return "", terror.Error(err, "Failed to unescape query")
	}
	u.RawQuery = decoded
	signedSignature := removeEncodePadding(signVoiceURL(u.String(), VoiceChatSecretKey))
	query.Add("signature", signedSignature)
	u.RawQuery = query.Encode()

	// remove percent encode
	decoded, err = url.QueryUnescape(query.Encode())
	if err != nil {
		gamelog.L.Error().Msg("Failed to decode url")
		return "", terror.Error(err, "Failed to unescape query")
	}
	u.RawQuery = decoded

	return u.String(), nil
}

// signs url with secret key
func signVoiceURL(url, secretKey string) string {
	h := hmac.New(sha1.New, []byte(secretKey))
	h.Write([]byte(url))
	return base64.RawURLEncoding.EncodeToString(h.Sum(nil))
}

func removeEncodePadding(s string) string {
	return strings.TrimRight(s, "=")
}
