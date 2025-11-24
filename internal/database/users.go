package database

import (
	"database/sql"
	"time"

	"context"
)

type UserModel struct {
	DB *sql.DB
}

type User struct {
	Id       int    `json:"id"`
	GoogleId string `json:"google_id"`
	Name     string `json:"name"`
	Password string `json:"-"`
}

func (um *UserModel) Insert(u *User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `INSERT INTO users (google_id, name, password) VALUES ($1, $2, $3) RETURNING id`

	return um.DB.QueryRowContext(ctx, query, u.GoogleId, u.Name, u.Password).Scan(&u.Id)
}

func (um *UserModel) Get(id int) (*User, error) {
	query := `SELECT id, google_id, name, password FROM users WHERE id = $1`
	return um.getUser(query, id)
}

func (um *UserModel) GetByGoogleID(googleID string) (*User, error) {
	query := `SELECT id, google_id, name, password FROM users WHERE google_id = $1`
	return um.getUser(query, googleID)
}

func (um *UserModel) getUser(query string, args ...any) (*User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var u User
	err := um.DB.QueryRowContext(ctx, query, args...).
		Scan(&u.Id, &u.GoogleId, &u.Name, &u.Password)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &u, nil
}
