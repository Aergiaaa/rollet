package database

import "database/sql"

type Models struct {
	Users  UserStore
	People PeopleStore
}

func NewModels(db *sql.DB) Models {
	return Models{
		Users:  &UserModel{DB: db},
		People: &PeopleModel{DB: db},
	}
}
