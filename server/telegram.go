package server

type Telegram interface {
	Notify(playerID string, mechID string, message string) error
	List()
	Insert()
	GenCode()
}
