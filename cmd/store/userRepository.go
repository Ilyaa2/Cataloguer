package store

import "Cataloguer/cmd/model"

type UserRepository interface {
	SaveUser(u *model.User) error
	FindByEmail(email string) (*model.User, error)
	FindByID(id int) (*model.User, error)
}
