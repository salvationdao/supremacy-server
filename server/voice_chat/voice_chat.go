package voice_chat

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/ws"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"golang.org/x/exp/slices"
	"net/url"
	"server"
	"server/battle"
	"server/db"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"strings"
	"time"
)

type VoiceChannel struct {
	Boston      []*boiler.VoiceStream
	Zaibatsu    []*boiler.VoiceStream
	RedMountain []*boiler.VoiceStream
}

type SignedPolicyURL struct {
	listenURL string
	sendURL   string
	expiredAt time.Time
}

var VoiceChatSecretKey string

func GetSignedPolicyURL(ownerID string) (*SignedPolicyURL, error) {
	baseURL := fmt.Sprintf("%s/%s", db.GetStrWithDefault(db.KeyOvenmediaAPIBaseUrl, "https://stream2.supremacy.game:8082"), ownerID)
	urlExpiryTime := db.GetIntWithDefault(db.KeyVoiceExpiryTimeHours, 2)
	signedPolicyURL := &SignedPolicyURL{}

	expiryTime := time.Now().Add(time.Hour * time.Duration(urlExpiryTime))
	signedPolicyURL.expiredAt = expiryTime

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

	signedPolicyURL.sendURL = sendURL
	signedPolicyURL.listenURL = listenURL

	return signedPolicyURL, nil
}

func UpdateVoiceChannel(warMachines []battle.WarMachine, arenaID string) error {
	_, err := boiler.VoiceStreams(
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

	for _, machineStream := range warMachines {
		if slices.Index(checkList, machineStream.OwnedByID) != -1 {
			continue
		}

		checkList = append(checkList, machineStream.OwnedByID)

		policyURL, err := GetSignedPolicyURL(machineStream.OwnedByID)
		if err != nil {
			gamelog.L.Error().Str("owner_id", machineStream.OwnedByID).Err(err).Msg("Failed to get signed policy url")
			continue
		}

		voiceStream := &boiler.VoiceStream{
			ArenaID:         arenaID,
			OwnerID:         machineStream.OwnedByID,
			FactionID:       machineStream.FactionID,
			IsActive:        true,
			SenderType:      boiler.VoiceSenderTypeMECH_OWNER,
			SendStreamURL:   policyURL.sendURL,
			ListenStreamURL: policyURL.listenURL,
			SessionExpireAt: policyURL.expiredAt,
		}

		err = voiceStream.Insert(gamedb.StdConn, boil.Infer())
		if err != nil {
			gamelog.L.Error().Str("owner_id", machineStream.OwnedByID).Err(err).Msg("Failed to insert voice stream")
			continue
		}

		switch machineStream.FactionID {
		case server.ZaibatsuFactionID:
			zaiChannel = append(zaiChannel, voiceStream)
		case server.RedMountainFactionID:
			rmChannel = append(rmChannel, voiceStream)
		case server.BostonCyberneticsFactionID:
			bostonChannel = append(bostonChannel, voiceStream)
		}
	}

	ps, err := boiler.Players(
		qm.Select(boiler.PlayerColumns.ID, boiler.PlayerColumns.FactionID),
		boiler.PlayerWhere.ID.IN(ws.TrackedIdents()),
		boiler.PlayerWhere.FactionID.IsNotNull(),
	).All(gamedb.StdConn)
	if err != nil {
		return err
	}

	for _, p := range ps {
		vcs := []*boiler.VoiceStream{}
		switch p.FactionID.String {
		case server.ZaibatsuFactionID:
			for _, zc := range zaiChannel {
				vc := &boiler.VoiceStream{
					ListenStreamURL: zc.ListenStreamURL,
				}

				if zc.OwnerID == p.ID {
					vc.SendStreamURL = zc.SendStreamURL
				}

				vcs = append(vcs, vc)
			}

			ws.PublishMessage(fmt.Sprintf("/secure/user/%s/faction_commander/%s", p.ID, server.ZaibatsuFactionID), "voice_stream", vcs)
		case server.RedMountainFactionID:
			for _, rc := range rmChannel {
				vc := &boiler.VoiceStream{
					ListenStreamURL: rc.ListenStreamURL,
				}

				if rc.OwnerID == p.ID {
					vc.SendStreamURL = rc.SendStreamURL
				}

				vcs = append(vcs, vc)
			}
			ws.PublishMessage(fmt.Sprintf("/secure/user/%s/faction_commander/%s", p.ID, server.RedMountainFactionID), "voice_stream", vcs)
		case server.BostonCyberneticsFactionID:
			for _, bc := range bostonChannel {
				vc := &boiler.VoiceStream{
					ListenStreamURL: bc.ListenStreamURL,
				}

				if bc.OwnerID == p.ID {
					vc.SendStreamURL = bc.SendStreamURL
				}

				vcs = append(vcs, vc)
			}
			ws.PublishMessage(fmt.Sprintf("/secure/user/%s/faction_commander/%s", p.ID, server.BostonCyberneticsFactionID), "voice_stream", vcs)
		}
	}

	return nil
}

func generateSignedURL(baseURL string, expiryTime time.Time, send bool) (string, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return "", terror.Error(err, "Failed to parse base url")
	}
	policy := fmt.Sprintf("{\"url_expire\":%d}", expiryTime.Unix())
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
