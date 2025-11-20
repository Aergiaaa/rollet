package database

import (
	"context"
	"database/sql"
	"time"
)

type PeopleModel struct {
	DB *sql.DB
}

type People struct {
	Name string `json:"name"`
	Role string `json:"role"`
	Team int    `json:"team"`
}

type PeopleData struct {
	Id     int `json:"id"`
	User   User
	People []People
}

func (pm *PeopleModel) Get() ([]*People, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `SELECT name, role, team FROM people WHERE id = $1`

	rows, err := pm.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	peoples := []*People{}

	for rows.Next() {
		var p People

		err := rows.Scan(&p.Name, &p.Role, &p.Team)
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
