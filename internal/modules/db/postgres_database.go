package db

import (
	"database/sql"
)

type PostgresConfig struct {
	Host string
	Port string
}

type DatabasePostgresImpl struct {
}

func (d *DatabasePostgresImpl) connect() error {
	//TODO implement me
	panic("implement me")
}

func (d *DatabasePostgresImpl) GetTables() ([]Table, error) {
	//TODO implement me
	panic("implement me")
}

func (d *DatabasePostgresImpl) Query(query string, args ...interface{}) (*sql.Rows, error) {
	//TODO implement me
	panic("implement me")
}

func newDatabasePostgresImpl(config PostgresConfig) Database {
	result := &DatabasePostgresImpl{}

	err := result.connect()
	if err != nil {
		panic(err)
	}
	return result
}
