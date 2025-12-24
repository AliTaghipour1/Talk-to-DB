package db

import (
	"database/sql"

	"github.com/AliTaghipour1/Talk-to_DB/pkg/repo"
)

type Database interface {
	connect() error
	GetTables() (Tables, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
}

type Column struct {
	Name     string
	DataType string
}

func (c Column) toRepositoryColumn() repo.Column {
	return repo.Column{Name: c.Name, DataType: c.DataType}
}

type Columns []Column

func (c Columns) toRepositoryColumnsList() []repo.Column {
	var result []repo.Column
	for _, column := range c {
		result = append(result, column.toRepositoryColumn())
	}
	return result
}

type Table struct {
	Name    string
	Columns Columns
}

func (t Table) toRepositoryTable() repo.Table {
	return repo.Table{
		Name:    t.Name,
		Columns: t.Columns.toRepositoryColumnsList(),
	}
}

type Tables []Table

func (t Tables) ToRepositoryTableList() []repo.Table {
	var result []repo.Table
	for _, table := range t {
		result = append(result, table.toRepositoryTable())
	}
	return result
}
