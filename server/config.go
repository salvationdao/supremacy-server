package server

type Config struct {
	CookieSecure          bool
	CookieKey             string
	EncryptTokens         bool
	EncryptTokensKey      string
	TokenExpirationDays   int
	TwitchUIHostURL       string
	ServerStreamKey       string
	PassportWebhookSecret string
	JwtKey                []byte
	Environment           string
	Address               string
	AuthCallbackURL       string
	AuthHangarCallbackURL string
}
