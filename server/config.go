package server

type Config struct {
	CookieSecure          bool
	EncryptTokens         bool
	EncryptTokensKey      string
	TokenExpirationDays   int
	TwitchUIHostURL       string
	ServerStreamKey       string
	PassportWebhookSecret string
	JwtKey                []byte
}
