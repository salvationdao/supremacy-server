package gameserver

type Config struct {
	CookieSecure        bool
	EncryptTokens       bool
	EncryptTokensKey    string
	TokenExpirationDays int
}
