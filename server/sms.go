package server

type SMS interface {
	SendSMS(to string, message string) error
	Lookup(number string) (string, error)
}
