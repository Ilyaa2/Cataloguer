package sqlstore

import (
	"Cataloguer/cmd/store"
	"database/sql"
	_ "github.com/lib/pq"
	"log"
)

type Sqlstore struct {
	connection        *sql.DB
	userRepository    *UserRepository
	messageRepository *MessageRepository
}

func New(url string) (*Sqlstore, error) {
	db, err := sql.Open("postgres", url)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		log.Println(err)
		return nil, err
	}
	return &Sqlstore{connection: db}, nil
}

func (s *Sqlstore) User() store.UserRepository {
	if s.userRepository == nil {
		s.userRepository = &UserRepository{
			SqlStore: s,
		}
	}
	return s.userRepository
}

func (s *Sqlstore) Message() store.MessageRepository {
	if s.messageRepository == nil {
		s.messageRepository = &MessageRepository{
			SqlStore: s,
		}
	}
	return s.messageRepository
}

/*
func wrapErrorFromDB(err error) error {
	if err == nil {
		return err
	}
	utf8Text, _ := charmap.Windows1251.NewDecoder().String(err.Error())
	return fmt.Errorf(utf8Text)
}

*/
