package server

type Telegram interface {
	Notify(code string, message string) error
}
