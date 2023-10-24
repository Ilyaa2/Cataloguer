package model

import (
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID             int    `json:"id,omitempty"`
	Email          string `json:"email"`
	Name           string `json:"name"`
	Password       string `json:"password,omitempty"`
	HashedPassword string `json:"-"`
}

func (u *User) BeforeCreate() error {
	res, err := encryptString(u.Password)
	if err != nil {
		return err
	}
	u.RemovePassword()
	u.HashedPassword = res
	return nil
}

func encryptString(s string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(s), bcrypt.MinCost)
	if err != nil {
		return "", err
	}

	return string(b), nil
}

func (u *User) RemovePassword() {
	u.Password = ""
}

func (u *User) IsPasswordCorrect(password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(u.HashedPassword), []byte(password)) == nil
}

func (u *User) ValidateUserFields() error {
	return validation.ValidateStruct(u,
		validation.Field(&u.Email, validation.Required, is.Email),
		validation.Field(&u.Password, validation.Required, validation.Length(4, 20)),
		validation.Field(&u.Name, validation.Required, validation.Length(1, 40)))
}

func (u *User) IsPayloadFieldsEmpty() bool {
	return u.Password == "" || u.Email == ""
}
