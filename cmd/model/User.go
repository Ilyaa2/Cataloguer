package model

import (
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
	u.Password = ""
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

// todo validateUserFields
func (u *User) ValidateUserFields() error {
	//должна быть проверка имени.
	return nil
}
