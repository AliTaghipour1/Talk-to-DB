package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"

	"github.com/AliTaghipour1/Talk-to_DB/pkg/repo"
)

type QueryResult struct {
	*sql.Rows
}

func (r QueryResult) Json() (string, error) {
	// Get column names
	columns, err := r.Columns()
	if err != nil {
		log.Printf("error getting columns: %v", err)
		return "", err
	}

	// Create a slice to hold all results
	var results []map[string]string

	// Iterate through rows
	for r.Next() {
		// Create a slice of interface{} to hold values
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		// Scan the row into value pointers
		if err := r.Scan(valuePtrs...); err != nil {
			log.Printf("error scanning row: %v", err)
			continue
		}

		// Create a map for this row, converting all values to strings
		rowMap := make(map[string]string)
		for i, col := range columns {
			val := values[i]
			var strVal string
			if val == nil {
				strVal = "null"
			} else {
				// Convert to string using fmt.Sprintf to handle all types
				strVal = fmt.Sprintf("%v", val)
			}
			rowMap[col] = strVal
		}
		results = append(results, rowMap)
	}

	// Check for errors from iterating over rows
	if err = r.Err(); err != nil {
		log.Printf("error iterating rows: %v", err)
		return "", err
	}

	// Marshal results to JSON
	jsonData, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		log.Printf("error marshaling to JSON: %v", err)
		return "", err
	}
	return string(jsonData), nil
}

type Database interface {
	connect() error
	GetTables() (Tables, error)
	Query(query string, args ...interface{}) (*QueryResult, error)
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
