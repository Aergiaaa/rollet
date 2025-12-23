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

	query := `
        SELECT pd.id, pd.created_at,
               p.id, p.name, p.role, p.team
        FROM people_data pd
        LEFT JOIN people p ON p.people_data_id = pd.id
        WHERE pd.user_id = $1
        ORDER BY pd.created_at DESC, p.team, p.role, p.name`
	rows, err := pm.DB.QueryContext(ctx, query, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	dataMap := make(map[int]*PeopleData)
	var order []int

	for rows.Next() {
		var pdID int
		var pdCreatedAt time.Time
		var pid sql.NullInt32
		var pname, prole sql.NullString
		var pteam sql.NullInt32

		err := rows.Scan(&pdID, &pdCreatedAt, &pid, &pname, &prole, &pteam)
		if err != nil {
			return nil, err
		}

		pd, ok := dataMap[pdID]
		if !ok {
			pd = &PeopleData{
				Id:        pdID,
				CreatedAt: pdCreatedAt,
			}
			dataMap[pdID] = pd
			order = append(order, pdID)
		}

		if pid.Valid {
			pd.People = append(pd.People, &People{
				Id:           int(pid.Int32),
				Name:         pname.String,
				Role:         prole.String,
				Team:         int(pteam.Int32),
				PeopleDataID: pdID,
			})
		}
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	peeps := make([]PeopleData, 0, len(order))

	for _, id := range order {
		peeps = append(peeps, *dataMap[id])
	}

	return peeps, nil
}

func (pm *PeopleModel) GetByUserId(userId, peopleId int) (*PeopleData, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
        SELECT pd.id, pd.created_at,
               p.id, p.name, p.role, p.team
        FROM people_data pd
        LEFT JOIN people p ON p.people_data_id = pd.id
        WHERE pd.user_id = $1 AND pd.id = $2
        ORDER BY p.team, p.role, p.name`
	rows, err := pm.DB.QueryContext(ctx, query, userId, peopleId)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	defer rows.Close()

	var pd *PeopleData
	for rows.Next() {
		var pdID int
		var pdCreated time.Time
		var pid sql.NullInt32
		var name, role sql.NullString
		var team sql.NullInt32

		if err := rows.Scan(&pdID, &pdCreated, &pid, &name, &role, &team); err != nil {
			return nil, err
		}

		if pd == nil {
			pd = &PeopleData{Id: pdID, CreatedAt: pdCreated}
		}

		if pid.Valid {
			pd.People = append(pd.People, &People{
				Id:           int(pid.Int32),
				Name:         name.String,
				Role:         role.String,
				Team:         int(team.Int32),
				PeopleDataID: pdID,
			})
		}
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	return pd, nil
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
		err = stmt.QueryRowContext(ctx, p.Name, p.Role, p.Team, userId, dataId).
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
