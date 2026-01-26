package bot

import (
	"fmt"
	"log"
	"slices"
	"strconv"
	"strings"

	"github.com/AliTaghipour1/Talk-to_DB/internal/config"
	"github.com/AliTaghipour1/Talk-to_DB/internal/modules/bot/messages"
	"github.com/AliTaghipour1/Talk-to_DB/internal/modules/database_handler"
	"github.com/AliTaghipour1/Talk-to_DB/pkg/bot_api"
	"github.com/AliTaghipour1/Talk-to_DB/pkg/repo"
	tgbotapi "github.com/ghiac/bale-bot-api"
)

type UpdateHandler struct {
	databaseHandler database_handler.DatabaseHandler
	sender          bot_api.BotApi
	botAPI          *tgbotapi.BotAPI
	allowedUserIds  []int64
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

	switch callback {
	default:
		if strings.HasPrefix(callback, "database-data-") {
			databaseID, err := strconv.Atoi(strings.TrimPrefix(callback, "database-data-"))
			if err != nil {
				log.Println("message - database callback parse failed:", err)
				return
			}
			u.handleChoosingDatabase(userID, databaseID)
		}
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
		u.handleStart(userID)
	case "/create_db":
		u.handleCreateDatabase(userID)
	case "/connect_postgres", "/connect_mysql", "/connect_cockroach":
		db := strings.TrimPrefix(text, "/connect_")
		u.handleSwitchDriver(db, userID)
	default:
		u.handleQuery(text, userID)
	}

	return
}

func (u *UpdateHandler) handleQuery(text string, userID int64) {
	result, err := u.databaseHandler.Query(text)
	if err != nil {
		log.Printf("error executing query: %v", err)
		u.sender.SendMessage(bot_api.Message{
			Text:   err.Error(),
			ChatId: userID,
		})
		return
	}

	u.sender.SendMessage(bot_api.Message{
		Text:   result,
		ChatId: userID,
	})
}

func (u *UpdateHandler) handleStart(userID int64) {
	databases, err := u.databaseHandler.GetDatabases()
	if err != nil {
		return
	}

	u.sender.SendMessage(bot_api.Message{
		Text:        "Choose databases:",
		ChatId:      userID,
		ReplyMarkup: messages.GenerateDatabaseButtons(createDatabaseMessageData(databases)),
	})
}

func (u *UpdateHandler) handleChoosingDatabase(userID int64, databaseID int) {
	err := u.databaseHandler.HandleChoosingDatabase(databaseID)
	if err != nil {
		return
	}

	u.sender.SendMessage(bot_api.Message{
		Text:   "کوئری رو بگو:",
		ChatId: userID,
	})
}

func createDatabaseMessageData(dbs []repo.Database) []messages.DatabaseData {
	var result []messages.DatabaseData
	for _, db := range dbs {
		result = append(result, messages.DatabaseData{
			Name: db.Name,
			ID:   db.ID,
		})
	}

	return result
}

func (u *UpdateHandler) handleSwitchDriver(db string, userID int64) {
	err := u.databaseHandler.SwitchDriver(db)
	if err != nil {
		return
	}

	u.sender.SendMessage(bot_api.Message{
		Text:   "Connected to " + db,
		ChatId: userID,
	})
}

func (u *UpdateHandler) handleCreateDatabase(userID int64) {
	database, err := u.databaseHandler.CreateDatabase()
	if err != nil {
		return
	}

	u.sender.SendMessage(bot_api.Message{
		Text:   "Created database with id:" + fmt.Sprint(database),
		ChatId: userID,
	})

}

func (u *UpdateHandler) isUserAllowedToUseBot(userId int64) bool {
	return slices.Contains(u.allowedUserIds, userId)
}

func NewBotUpdateHandler(sender bot_api.BotApi, botAPI *tgbotapi.BotAPI, serviceConfig *config.TalkToDBConfig) *UpdateHandler {
	result := &UpdateHandler{
		botAPI:         botAPI,
		sender:         sender,
		allowedUserIds: serviceConfig.AllowedUserIds,
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
