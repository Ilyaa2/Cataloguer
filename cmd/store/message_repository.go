package store

import "Cataloguer/cmd/model"

type MessageRepository interface {
	GetMessagesByUserId(userId int) ([]*model.Message, error)
	GetMessageById(messageId int) (*model.Message, error)
	GetMessageByName(messageName string) (*model.Message, error)
	CreateMessage(message *model.Message, userId int) error
	HasRightsOnMessageById(userId int, messageId int) bool
	HasRightsOnMessageByName(userId int, messageName string) bool
	DeleteMessageById(id int)
	DeleteMessageByName(name string)
}
