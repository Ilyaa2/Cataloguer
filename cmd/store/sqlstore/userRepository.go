package sqlstore

import "Cataloguer/cmd/model"

type UserRepository struct {
	SqlStore *Sqlstore
}

func (r *UserRepository) SaveUser(u *model.User) error {
	if err := u.BeforeCreate(); err != nil {
		return wrapErrorFromDB(err)
	}
	item := r.SqlStore.connection.QueryRow(
		"INSERT INTO users(name, email, encrypted_password) VALUES($1, $2, $3) returning id",
		u.Name, u.Email, u.HashedPassword,
	)
	var id int
	err := item.Scan(&id)
	if err != nil {
		return wrapErrorFromDB(err)
	}
	u.RemovePassword()
	u.ID = id
	return nil
}

//TODO ВОЗМОЖНО ЭТО ЛУЧШЕ ДЕЛАТЬ ЧЕРЕЗ PREPARED STATEMENT

func (r *UserRepository) FindByEmail(email string) (*model.User, error) {
	// может быть потребуется ввести все столбцы в правильном порядке в строке запроса.
	row := r.SqlStore.connection.QueryRow("SELECT * from users where email = $1", email)
	u := &model.User{}
	if err := row.Scan(&u.ID, &u.Name, &u.Email, &u.HashedPassword); err != nil {
		return nil, wrapErrorFromDB(err)
	}
	return u, nil
}

func (r *UserRepository) FindByID(id int) (*model.User, error) {
	row := r.SqlStore.connection.QueryRow("SELECT * from users where id = $1", id)
	u := &model.User{}
	if err := row.Scan(&u.ID, &u.Name, &u.Email, &u.HashedPassword); err != nil {
		return nil, wrapErrorFromDB(err)
	}
	return u, nil
}
