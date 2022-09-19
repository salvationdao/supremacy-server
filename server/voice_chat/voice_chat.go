package voice_chat

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"github.com/ninja-software/terror/v2"
	"net/url"
	"server/db"
	"server/db/boiler"
	"server/gamelog"
	"strings"
	"time"
)

type VoiceChannel struct {
	Boston      map[string]*boiler.VoiceStream
	Zaibatsu    map[string]*boiler.VoiceStream
	RedMountain map[string]*boiler.VoiceStream
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
