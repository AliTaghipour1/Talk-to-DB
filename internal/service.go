package internal

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/AliTaghipour1/Talk-to_DB/internal/config"
	"github.com/AliTaghipour1/Talk-to_DB/internal/modules/ai"
	"github.com/AliTaghipour1/Talk-to_DB/internal/modules/bot"
	"github.com/AliTaghipour1/Talk-to_DB/internal/modules/database_handler"
	db2 "github.com/AliTaghipour1/Talk-to_DB/internal/modules/db"
	"github.com/AliTaghipour1/Talk-to_DB/pkg/bot_api"
	"github.com/AliTaghipour1/Talk-to_DB/pkg/repo"
	tgbotapi "github.com/ghiac/bale-bot-api"
)

type Service struct {
	databases map[config.Driver]db2.Database //maybe it is better to use name instead of driver
}

func NewService() *Service {
	return &Service{
		databases: make(map[config.Driver]db2.Database),
	}
}

func (s *Service) Run() {
	serviceConfig := config.NewTalkToDbConfig()
	configJsonText, _ := json.MarshalIndent(serviceConfig, "", "    ")
	log.Println("config:", string(configJsonText))

	s.createDatabases(serviceConfig.Databases)
	databaseRepo := repo.NewDatabaseRepoMapImpl("pkg/repo/data.json")
	s.runBot(serviceConfig, databaseRepo)
}

func (s *Service) createDatabases(dbs []config.Database) {
	for _, db := range dbs {
		cfg, driver := convertDatabaseConfigModel(db)
		database, err := db2.NewDatabase(cfg, driver)
		if err != nil {
			panic(fmt.Errorf("failed to create [%s] database: %v", db.Driver, err))
		}
		s.databases[db.Driver] = database
	}
}

func (s *Service) runBot(serviceConfig *config.TalkToDBConfig, databaseRepo repo.DatabaseRepo) {
	botApi := getBotApi(serviceConfig.CliBot.Token, serviceConfig.DebugMode)
	sender := bot_api.NewSenderBot(botApi)

	dbHandler := database_handler.NewDatabaseHandler(serviceConfig.AllowedUserIds, s.databases, databaseRepo, ai.NewAIModule(serviceConfig.AvalAi.ApiKey))
	bot.NewBotUpdateHandler(dbHandler, sender, botApi, serviceConfig).Start()
}

func getBotApi(token string, debugMode bool) *tgbotapi.BotAPI {
	botApi, err := tgbotapi.NewBaleBotAPI(token)
	if err != nil {
		panic(err)
	}
	botApi.Client.Timeout = 10 * time.Second
	botApi.Debug = debugMode
	botApi.Buffer = 5000

	log.Printf("Authorized on account %s", botApi.Self.UserName)

	return botApi
}

func convertDatabaseConfigModel(database config.Database) (db2.DatabaseConfig, db2.Driver) {
	var driver db2.Driver
	switch database.Driver {
	case config.Postgres:
		driver = db2.Postgres
	case config.MySQL:
		driver = db2.MySQL
	case config.Cockroach:
		driver = db2.Cockroach
	}

	return db2.DatabaseConfig{
		Host:     database.Host,
		Port:     database.Port,
		User:     database.User,
		Password: database.Pass,
		Database: database.Name,
		SSLMode:  "disable",
		Schema:   "public",
	}, driver
}
