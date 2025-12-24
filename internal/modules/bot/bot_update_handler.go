package bot

import (
	"fmt"
	"log"
	"slices"
	"strings"

	"github.com/AliTaghipour1/Talk-to_DB/internal/config"
	db2 "github.com/AliTaghipour1/Talk-to_DB/internal/modules/db"
	"github.com/AliTaghipour1/Talk-to_DB/pkg/bot_api"
	"github.com/AliTaghipour1/Talk-to_DB/pkg/repo"
	tgbotapi "github.com/ghiac/bale-bot-api"
)

type UpdateHandler struct {
	sender         bot_api.BotApi
	botAPI         *tgbotapi.BotAPI
	allowedUserIds []int64
	databases      map[config.Driver]db2.Database
	currentDriver  config.Driver
	databaseRepo   repo.DatabaseRepo
}

func (u *UpdateHandler) handleUpdate(update *tgbotapi.Update) {
	if update.CallbackQuery != nil {
		chat := update.CallbackQuery.Message.Chat
		if chat.Type != "private" {
			u.handleNonPrivateUpdate(chat.ID)
			return
		}
		u.handleCallback(update)
	} else if update.Message != nil {
		chat := update.Message.Chat
		if chat.Type != "private" {
			u.handleNonPrivateUpdate(chat.ID)
			return
		}

		u.handleMessage(update)
	}
}

func (u *UpdateHandler) handleNonPrivateUpdate(chatId int64) {
	leaveChat, err := u.botAPI.LeaveChat(tgbotapi.ChatConfig{
		ChatID: chatId,
	})
	if err != nil {
		log.Println("message - leave chat failed:", err)
		return
	}
	log.Println("message - leave chat success:", leaveChat)
	return

}

func (u *UpdateHandler) handleCallback(update *tgbotapi.Update) {
	callback := update.CallbackQuery.Data
	userID := int64(update.CallbackQuery.From.ID)
	callbackQueryId := update.CallbackQuery.ID

	if !u.isUserAllowedToUseBot(userID) {
		return
	}

	u.sender.SendMessage(bot_api.Message{
		Text:   callback + callbackQueryId,
		ChatId: userID,
	})

	return
}

func (u *UpdateHandler) handleMessage(update *tgbotapi.Update) {
	userID := int64(update.Message.From.ID)

	if !u.isUserAllowedToUseBot(userID) {
		return
	}

	text := update.Message.Text
	switch text {
	case "/start":

	case "/create_db":
		u.handleCreateDatabase(userID)

	case "/connect_postgres", "/connect_mysql":
		db := strings.TrimPrefix(text, "/connect_")
		u.handleSwitchDriver(db, userID)
	default:
		u.sender.SendMessage(bot_api.Message{
			Text:   update.Message.Text,
			ChatId: userID,
		})
	}

	return
}

func (u *UpdateHandler) handleSwitchDriver(db string, userID int64) {
	switch db {
	case "postgres":
		u.currentDriver = config.Postgres
	case "mysql":
		u.currentDriver = config.MySQL
	}

	u.sender.SendMessage(bot_api.Message{
		Text:   "Connected to " + db,
		ChatId: userID,
	})
}

func (u *UpdateHandler) handleCreateDatabase(userID int64) {
	if u.currentDriver == "" {
		return
	}

	tables, err := u.databases[u.currentDriver].GetTables()
	if err != nil {
		return
	}

	db := &repo.Database{
		Name:   string(u.currentDriver),
		Tables: tables.ToRepositoryTableList(),
	}
	databaseID, err := u.databaseRepo.CreateNewDatabase(db)
	if err != nil {
		return
	}

	u.sender.SendMessage(bot_api.Message{
		Text:   "Created database with id:" + fmt.Sprint(databaseID),
		ChatId: userID,
	})

}

func (u *UpdateHandler) isUserAllowedToUseBot(userId int64) bool {
	return slices.Contains(u.allowedUserIds, userId)
}

func NewBotUpdateHandler(sender bot_api.BotApi, botAPI *tgbotapi.BotAPI, databases map[config.Driver]db2.Database, serviceConfig *config.TalkToDBConfig) *UpdateHandler {
	result := &UpdateHandler{
		botAPI:         botAPI,
		sender:         sender,
		allowedUserIds: serviceConfig.AllowedUserIds,
		databases:      databases,
		currentDriver:  "",
	}

	return result
}

func (u *UpdateHandler) Start() {
	updatesChan, err := u.botAPI.GetUpdatesChan(tgbotapi.NewUpdate(0))
	if err != nil {
		panic(err)
	}

	for update := range updatesChan {
		u.handleUpdate(&update)
	}
}
