package database_handler

import (
	"errors"
	"fmt"

	"github.com/AliTaghipour1/Talk-to_DB/internal/config"
	"github.com/AliTaghipour1/Talk-to_DB/internal/modules/ai"
	db2 "github.com/AliTaghipour1/Talk-to_DB/internal/modules/db"
	"github.com/AliTaghipour1/Talk-to_DB/pkg/repo"
	tgbotapi "github.com/ghiac/bale-bot-api"
)

type DatabaseHandler struct {
	botAPI          *tgbotapi.BotAPI
	allowedUserIds  []int64
	databases       map[config.Driver]db2.Database
	currentDriver   config.Driver
	currentDatabase *repo.Database
	databaseRepo    repo.DatabaseRepo
	aiModule        *ai.AIModule
}

var (
	ErrEmptyDriver  = errors.New("no available database driver")
	ErrNotConnected = errors.New("not connected")
)

func (d *DatabaseHandler) HandleChoosingDatabase(databaseID int) error {
	if d.currentDriver == "" {
		return ErrEmptyDriver
	}

	database, err := d.databaseRepo.GetDatabase(databaseID)
	if err != nil {
		return err
	}

	d.currentDatabase = &database

	return nil
}

func (d *DatabaseHandler) GetDatabases() ([]repo.Database, error) {
	if d.currentDriver == "" {
		return nil, ErrEmptyDriver
	}

	databases, err := d.databaseRepo.GetAllDatabases()
	if err != nil {
		return nil, err
	}

	return databases, nil
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
	if d.currentDatabase == nil {
		return "", ErrNotConnected
	}

	query := d.aiModule.GetQuery(d.currentDatabase.Scheme(), text)

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
