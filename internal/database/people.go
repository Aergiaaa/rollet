package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type PeopleModel struct {
	DB *sql.DB
}

type People struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
	Role string `json:"role"`
	Team int    `json:"team"`
}

type PeopleData struct {
	Id     int `json:"id"`
	User   User
	People []People
}

func (pm *PeopleModel) GetAllbyUserId(userId int) ([]*People, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `SELECT id, name, role, team FROM people WHERE user_id = $1 ORDER BY role, name`
	rows, err := pm.DB.QueryContext(ctx, query, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	peoples := []*People{}

	for rows.Next() {
		var p People

		err := rows.Scan(&p.Id, &p.Name, &p.Role, &p.Team)
		if err != nil {
			return nil, err
		}

		peoples = append(peoples, &p)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return peoples, nil
}

func (pm *PeopleModel) Save(userId int, people []*People) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tx, err := pm.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	query := `INSERT INTO people (name, role, team, user_id) VALUES ($1, $2, $3, $4) RETURNING id`
	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, p := range people {
		err = stmt.QueryRowContext(ctx, p.Name, p.Role, p.Team, userId).
			Scan(&p.Id)
		if err != nil {
			return fmt.Errorf("failed to insert people: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
