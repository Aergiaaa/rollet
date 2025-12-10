package database

import (
	"database/sql"
	"time"

	"context"
)

type UserStore interface {
	Insert(u *User) error
	Get(id int) (*User, error)
	GetByEmail(Email string) (*User, error)
	GetByName(name string) (*User, error)
}

type UserModel struct {
	DB *sql.DB
}

type User struct {
	Id       int    `json:"id"`
	Email    string `json:"email"`
	GoogleID string `json:"google_id,omitempty"`
	Name     string `json:"name"`
	Password string `json:"-"`
}

var _ UserStore = (*UserModel)(nil)

func (um *UserModel) Insert(u *User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `INSERT INTO users (email, google_id, name, password) VALUES ($1, $2, $3, $4) RETURNING id`

	return um.DB.QueryRowContext(ctx, query, u.Email, u.GoogleID, u.Name, u.Password).Scan(&u.Id)
}

func (um *UserModel) Get(id int) (*User, error) {
	query := `SELECT id, email, google_id, name, password FROM users WHERE id = $1`
	return um.getUser(query, id)
}

func (um *UserModel) GetByEmail(email string) (*User, error) {
	query := `SELECT id, email, google_id, name, password FROM users WHERE email = $1`
	return um.getUser(query, email)
}

func (um *UserModel) GetByName(name string) (*User, error) {
	query := `SELECT id, email, google_id, name, password FROM users WHERE name = $1`
	return um.getUser(query, name)
}

func (um *UserModel) getUser(query string, args ...any) (*User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var u User
	err := um.DB.QueryRowContext(ctx, query, args...).
		Scan(&u.Id, &u.Email, &u.Name, &u.Password)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &u, nil
}
