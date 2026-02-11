package database_handler

import (
	"errors"
	"fmt"

	"github.com/AliTaghipour1/Talk-to_DB/internal/config"
	"github.com/AliTaghipour1/Talk-to_DB/internal/modules/ai"
	db2 "github.com/AliTaghipour1/Talk-to_DB/internal/modules/db"
	"github.com/AliTaghipour1/Talk-to_DB/pkg/repo"
)

type DatabaseHandler struct {
	allowedUserIds    []int64
	databases         map[config.Driver]db2.Database
	currentDriver     config.Driver
	currentDatabase   *repo.Database
	currentDatabaseID *int
	databaseRepo      repo.DatabaseRepo
	aiModule          *ai.AIModule
}

func NewDatabaseHandler(allowedUserIds []int64, databases map[config.Driver]db2.Database, databaseRepo repo.DatabaseRepo,
	aiModule *ai.AIModule) *DatabaseHandler {
	return &DatabaseHandler{
		allowedUserIds: allowedUserIds,
		databases:      databases,
		databaseRepo:   databaseRepo,
		aiModule:       aiModule,
	}
}

var (
	ErrEmptyDriver  = errors.New("no available database driver")
	ErrNotConnected = errors.New("not connected")
)

func (d *DatabaseHandler) HandleChoosingDatabase(databaseID int) error {
	if d.currentDriver == "" {
		return ErrEmptyDriver
	}

	_, err := d.databaseRepo.GetDatabase(databaseID)
	if err != nil {
		return err
	}

	d.currentDatabaseID = &databaseID

	return nil
}

func (d *DatabaseHandler) GetDatabases() ([]Database, error) {
	if d.currentDriver == "" {
		return nil, ErrEmptyDriver
	}

	databases, err := d.databaseRepo.GetAllDatabases()
	if err != nil {
		return nil, err
	}

	return convertRepoDatabasesToModuleModel(databases), nil
}

func (d *DatabaseHandler) GetCurrentDatabase() (Database, error) {
	if d.currentDatabaseID == nil {
		return Database{}, ErrEmptyDriver
	}

	currentDatabase, err := d.databaseRepo.GetDatabase(*d.currentDatabaseID)
	if err != nil {
		return Database{}, err
	}

	db := convertRepoDatabasesToModuleModel([]repo.Database{currentDatabase})
	return db[0], nil
}

func (d *DatabaseHandler) CreateDatabase() (int, error) {
	if d.currentDriver == "" {
		return 0, ErrEmptyDriver
	}

	tables, err := d.databases[d.currentDriver].GetTables()
	if err != nil {
		return 0, err
	}

	db := &repo.Database{
		Name:   string(d.currentDriver),
		Tables: tables.ToRepositoryTableList(),
	}
	databaseID, err := d.databaseRepo.CreateNewDatabase(db)
	if err != nil {
		return 0, err
	}

	return databaseID, nil

}

func (d *DatabaseHandler) SwitchDriver(db string) error {
	switch db {
	case "postgres":
		d.currentDriver = config.Postgres
	case "mysql":
		d.currentDriver = config.MySQL
	case "cockroach":
		d.currentDriver = config.Cockroach
	default:
		return errors.New("unknown database driver")
	}

	return nil
}

func (d *DatabaseHandler) Query(text string) (string, error) {
	if d.currentDatabaseID == nil {
		return "", ErrNotConnected
	}

	currentDatabase, err := d.databaseRepo.GetDatabase(*d.currentDatabaseID)
	if err != nil {
		return "", err
	}

	database := convertRepoDatabaseToModuleModel(currentDatabase)

	query := d.aiModule.GetQuery(database.Scheme(), text)

	rows, err := d.databases[d.currentDriver].Query(query)
	if err != nil {
		return "", fmt.Errorf(`error executing query: %v`, err)
	}
	defer rows.Close()

	result, err := rows.Json()
	if err != nil {
		return "", fmt.Errorf("Error converting result to json: %v", err)
	}

	return result, nil
}

func (d *DatabaseHandler) SetDescription(tableName string, columnName *string, description string) error {
	if d.currentDatabaseID == nil {
		return ErrNotConnected
	}
	currentDatabase, err := d.databaseRepo.GetDatabase(*d.currentDatabaseID)
	if err != nil {
		return err
	}

	table, found := currentDatabase.GetTableByName(tableName)
	if !found {
		return errors.New("table not found")
	}
	filedType := repo.TableFieldType
	fieldID := table.ID
	if columnName != nil {
		filedType = repo.ColumnFieldType
		column, columnFound := table.GetColumnByName(*columnName)
		if !columnFound {
			return errors.New("column not found")
		}
		fieldID = column.ID
	}

	err = d.databaseRepo.SetDescription(currentDatabase.ID, description, fieldID, filedType)
	if err != nil {
		return err
	}
	return nil
}
