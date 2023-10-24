package sqlstore

import (
	"Cataloguer/cmd/model"
	"errors"
)

type MessageRepository struct {
	SqlStore *Sqlstore
}

func (m *MessageRepository) GetMessagesByUserId(userId int) ([]*model.Message, error) {
	rows, err := m.SqlStore.connection.Query("SELECT id, name, type, time, path FROM messages WHERE user_id = $1", userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var allMessages []*model.Message
	flag := false
	for rows.Next() {
		flag = true
		message := model.Message{}
		err := rows.Scan(&message.Id, &message.Name, &message.Type, &message.DateTime, &message.Path)
		if err != nil {
			return nil, err
		}
		allMessages = append(allMessages, &message)
	}
	if !flag {
		return nil, errors.New("no files found")
	}
	return allMessages, nil
}

func (m *MessageRepository) GetMessageById(messageId int) (*model.Message, error) {
	row := m.SqlStore.connection.QueryRow("SELECT id, name, type, time, path FROM messages WHERE id = $1", messageId)
	message := model.Message{}
	err := row.Scan(&message.Id, &message.Name, &message.Type, &message.DateTime, &message.Path)
	if err != nil {
		return nil, err
	}
	return &message, nil
}

func (m *MessageRepository) GetMessageByName(messageName string) (*model.Message, error) {
	row := m.SqlStore.connection.QueryRow("SELECT id, name, type, time, path FROM messages WHERE name = $1", messageName)
	message := model.Message{}
	err := row.Scan(&message.Id, &message.Name, &message.Type, &message.DateTime, &message.Path)
	if err != nil {
		return nil, err
	}
	return &message, nil
}

func (m *MessageRepository) HasRightsOnMessageById(userId int, messageId int) bool {
	row := m.SqlStore.connection.QueryRow("SELECT id FROM messages WHERE user_id = $1 AND id = $2", userId, messageId)
	var id int
	err := row.Scan(&id)
	if err != nil || id != messageId {
		return false
	}
	return true
}

func (m *MessageRepository) DeleteMessageById(id int) {
	_ = m.SqlStore.connection.QueryRow("DELETE FROM messages WHERE id = $1", id)
}

func (m *MessageRepository) HasRightsOnMessageByName(userId int, messageName string) bool {
	row := m.SqlStore.connection.QueryRow("SELECT name FROM messages WHERE user_id = $1 AND name = $2", userId, messageName)
	var msgName string
	err := row.Scan(&msgName)
	if err != nil || msgName != messageName {
		return false
	}
	return true
}

func (m *MessageRepository) DeleteMessageByName(name string) {
	_ = m.SqlStore.connection.QueryRow("DELETE FROM messages WHERE name = $1", name)
}

func (m *MessageRepository) CreateMessage(message *model.Message, userId int) error {
	row := m.SqlStore.connection.QueryRow("INSERT INTO messages(user_id, name, type, time, path) VALUES($1,$2,$3,$4,$5) returning id",
		userId, message.Name, message.Type, message.DateTime, message.Path)
	err := row.Scan(&(message.Id))
	if err != nil {
		return err
	}
	return nil
}
