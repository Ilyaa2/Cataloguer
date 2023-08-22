package sqlstore

import (
	"Cataloguer/cmd/store"
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"golang.org/x/text/encoding/charmap"
)

type Sqlstore struct {
	//тут должны хранится подключение (может пул)
	connection     *sql.DB
	userRepository *UserRepository
}

func New(url string) (*Sqlstore, error) {
	db, err := sql.Open("postgres", "user=postgres password=Tylpa31 dbname=cataloguer_test sslmode=disable")
	if err != nil {
		return nil, wrapErrorFromDB(err)
	}
	if err = db.Ping(); err != nil {
		fmt.Print(wrapErrorFromDB(err))
		return nil, wrapErrorFromDB(err)
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

func wrapErrorFromDB(err error) error {
	if err == nil {
		return err
	}
	utf8Text, _ := charmap.Windows1251.NewDecoder().String(err.Error())
	return fmt.Errorf(utf8Text)
}
