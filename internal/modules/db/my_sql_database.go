package db

import "database/sql"

type MySqlConfig struct {
	Host string
	Port string
}

type DatabaseMySqlImpl struct {
}

func (d *DatabaseMySqlImpl) connect() error {
	//TODO implement me
	panic("implement me")
}

func (d *DatabaseMySqlImpl) GetTables() ([]Table, error) {
	//TODO implement me
	panic("implement me")
}

func (d *DatabaseMySqlImpl) Query(query string, args ...interface{}) (*sql.Rows, error) {
	//TODO implement me
	panic("implement me")
}

func newDatabaseMySqlImpl(config MySqlConfig) Database {
	result := &DatabaseMySqlImpl{}

	err := result.connect()
	if err != nil {
		panic(err)
	}
	return result
}
