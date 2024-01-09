package database

import "database/sql"

type model struct {
	db *sql.DB
}

func NewModel(db *sql.DB) *model {
	return &model{db: db}
}