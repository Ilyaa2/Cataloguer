package store

type MessageRepository interface {
	GetMessage()
	CreateMessage()
	UpdateMessage()
	DeleteMessage(id int) error
}
