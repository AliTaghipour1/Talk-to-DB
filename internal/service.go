package internal

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/AliTaghipour1/Talk-to_DB/internal/config"
	db2 "github.com/AliTaghipour1/Talk-to_DB/internal/modules/db"
	"github.com/AliTaghipour1/Talk-to_DB/pkg/bot_api"
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
	runBot(serviceConfig)
}

func (s *Service) createDatabases(dbs []config.Database) {
	for _, db := range dbs {
		var database db2.Database
		var err error
		switch db.Driver {
		case config.Postgres:
			database, err = db2.NewDatabasePostgresImpl(db2.PostgresConfig{
				Host:     db.Host,
				Port:     db.Port,
				User:     db.User,
				Password: db.Pass,
				DBName:   db.Name,
			})
		case config.MySQL:
			database, err = db2.NewDatabaseMySqlImpl(db2.MySqlConfig{
				Host:     db.Host,
				Port:     db.Port,
				User:     db.User,
				Password: db.Pass,
				Database: db.Name,
			})
		}
		if err != nil {
			panic(fmt.Errorf("failed to create [%s] database: %v", db.Driver, err))
		}
		s.databases[db.Driver] = database
	}
}

func runBot(serviceConfig *config.TalkToDBConfig) {
	botApi := getBotApi(serviceConfig.CliBot.Token, serviceConfig.DebugMode)

	_ = bot_api.NewSenderBot(botApi)

	//TODO
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
