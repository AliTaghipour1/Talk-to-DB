package db

import "database/sql"

type Database interface {
	connect() error
	GetTables() ([]Table, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
}

type Column struct {
	Name     string
	DataType string
}

type Table struct {
	Name    string
	Columns []Column
}
