package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type PeopleStore interface {
	GetAllbyUserId(userId int) ([]PeopleData, error)
	GetByUserId(userId, peopleId int) (*PeopleData, error)
	DeleteByUserId(userId, peopleId int) error
	Save(userId int, people []*People) error
}

type PeopleModel struct {
	DB *sql.DB
}

type People struct {
	Id           int    `json:"id"`
	Name         string `json:"name"`
	Role         string `json:"role"`
	Team         int    `json:"team"`
	PeopleDataID int    `json:"-"`
}

type PeopleData struct {
	Id        int
	User      *User
	People    []*People
	CreatedAt time.Time
}

var _ PeopleStore = (*PeopleModel)(nil)

func (pm *PeopleModel) GetAllbyUserId(userId int) ([]PeopleData, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `SELECT id, created_at FROM people_data WHERE user_id = $1 ORDER BY created_at DESC`
	rows, err := pm.DB.QueryContext(ctx, query, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var peeps []PeopleData

	for rows.Next() {
		var pd PeopleData

		err := rows.Scan(&pd.Id, &pd.CreatedAt)
		if err != nil {
			return nil, err
		}

		people, err := pm.getPeopleByDataID(ctx, pd.Id)

		if err != nil {
			return nil, err
		}

		pd.People = people
		peeps = append(peeps, pd)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return peeps, nil
}

func (pm *PeopleModel) GetByUserId(userId, peopleId int) (*PeopleData, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var pd PeopleData

	query := `SELECT id, created_at FROM people_data WHERE id = $1 AND user_id = $2`
	err := pm.DB.QueryRowContext(ctx, query, peopleId, userId).
		Scan(&pd.Id, &pd.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	people, err := pm.getPeopleByDataID(ctx, pd.Id)
	if err != nil {
		return nil, err
	}

	pd.People = people

	return &pd, nil
}

func (pm *PeopleModel) DeleteByUserId(userId, peopleId int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `DELETE FROM people_data WHERE id = $1 AND user_id = $2`
	_, err := pm.DB.ExecContext(ctx, query, peopleId, userId)
	if err != nil {
		return err
	}

	return nil
}

func (pm *PeopleModel) Save(userId int, people []*People) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tx, err := pm.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	var dataId int

	query := `INSERT INTO people_data (user_id) VALUES ($1) RETURNING id`
	err = tx.QueryRowContext(ctx, query, userId).
		Scan(&dataId)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}

	query = `INSERT INTO people (name, role, team, user_id, people_data_id) VALUES ($1, $2, $3, $4, $5) RETURNING id`
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
		p.PeopleDataID = dataId
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (pm *PeopleModel) getPeopleByDataID(ctx context.Context, dataID int) ([]*People, error) {
	query := `SELECT id, name, role, team FROM people WHERE people_data_id = $1 ORDER BY role, name`
	rows, err := pm.DB.QueryContext(ctx, query, dataID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var people []*People
	for rows.Next() {
		var p People
		err := rows.Scan(&p.Id, &p.Name, &p.Role, &p.Team)
		if err != nil {
			return nil, err
		}
		p.PeopleDataID = dataID
		people = append(people, &p)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return people, nil
}
