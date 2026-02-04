package database_handler

import (
	"encoding/json"

	"github.com/AliTaghipour1/Talk-to_DB/pkg/repo"
)

type Table struct {
	Name        string
	Description string
	Columns     []Column
}

type Column struct {
	Name        string
	DataType    string
	Description string
}

type Database struct {
	ID          int
	Name        string
	Description string
	Tables      []Table
}

func (s Database) GetTableByName(name string) (Table, bool) {
	for _, table := range s.Tables {
		if table.Name == name {
			return table, true
		}
	}
	return Table{}, false
}

func (s Table) GetColumnByName(name string) (Column, bool) {
	for _, column := range s.Columns {
		if column.Name == name {
			return column, true
		}
	}
	return Column{}, false
}

func (s Database) Scheme() string {
	indent, err := json.MarshalIndent(s, "", "\t")
	if err != nil {
		panic(err)
	}
	return string(indent)
}

func convertRepoDatabasesToModuleModel(databases []repo.Database) []Database {
	var result []Database
	for _, database := range databases {
		result = append(result, convertRepoDatabaseToModuleModel(database))
	}

	return result
}

func convertRepoDatabaseToModuleModel(database repo.Database) Database {
	return Database{
		ID:          database.ID,
		Name:        database.Name,
		Description: database.Description,
		Tables:      convertRepoTableToModuleModel(database.Tables),
	}
}

func convertRepoTableToModuleModel(tables []repo.Table) []Table {
	var result []Table
	for _, table := range tables {
		result = append(result, Table{
			Name:        table.Name,
			Description: table.Description,
			Columns:     convertRepoColumnToModuleModel(table.Columns),
		})
	}
	return result
}

func convertRepoColumnToModuleModel(columns []repo.Column) []Column {
	var result []Column
	for _, column := range columns {
		result = append(result, Column{
			Name:        column.Name,
			DataType:    column.DataType,
			Description: column.Description,
		})
	}
	return result
}
