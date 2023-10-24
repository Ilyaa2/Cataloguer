package sqlstore

import "Cataloguer/cmd/model"

type UserRepository struct {
	SqlStore *Sqlstore
}

func (r *UserRepository) Save(u *model.User) error {
	if err := u.BeforeCreate(); err != nil {
		return err
	}
	item := r.SqlStore.connection.QueryRow(
		"INSERT INTO users(name, email, encrypted_password) VALUES($1, $2, $3) returning id",
		u.Name, u.Email, u.HashedPassword,
	)
	err := item.Scan(&(u.ID))
	if err != nil {
		return err
	}
	return nil
}

func (r *UserRepository) FindByEmail(email string) (*model.User, error) {
	row := r.SqlStore.connection.QueryRow("SELECT * from users where email = $1", email)
	u := &model.User{}
	if err := row.Scan(&u.ID, &u.Name, &u.Email, &u.HashedPassword); err != nil {
		return nil, err
	}
	return u, nil
}

func (r *UserRepository) FindByID(id int) (*model.User, error) {
	row := r.SqlStore.connection.QueryRow("SELECT * from users where id = $1", id)
	u := &model.User{}
	if err := row.Scan(&u.ID, &u.Name, &u.Email, &u.HashedPassword); err != nil {
		return nil, err
	}
	return u, nil
}
