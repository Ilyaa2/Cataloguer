package sqlstore

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetMessage(t *testing.T) {
	sqlstore, config := TestSqlStore(t)

	message, err := sqlstore.Message().GetMessageById(4)
	if err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(config.BasePath, filepath.FromSlash(message.Path))
	if err != nil {
		t.Fatal(err)
	}
	file, err := os.Open(path)
	defer file.Close()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(file.Name())
}

func TestGetMessagesByUserId(t *testing.T) {
	sqlstore, _ := TestSqlStore(t)
	allMessages, err := sqlstore.Message().GetMessagesByUserId(76)
	if err != nil {
		t.Fatal(err)
	}
	for _, message := range allMessages {
		t.Log(message.Id, message.DateTime, message.Path, message.Name, message.Type)
	}
}
