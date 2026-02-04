package bot

import (
	"fmt"
	"log"
	"slices"
	"strconv"
	"strings"
	"sync"

	"github.com/AliTaghipour1/Talk-to_DB/internal/config"
	"github.com/AliTaghipour1/Talk-to_DB/internal/modules/bot/messages"
	"github.com/AliTaghipour1/Talk-to_DB/internal/modules/database_handler"
	"github.com/AliTaghipour1/Talk-to_DB/pkg/bot_api"
	tgbotapi "github.com/ghiac/bale-bot-api"
)

type UpdateHandler struct {
	databaseHandler  *database_handler.DatabaseHandler
	sender           bot_api.BotApi
	botAPI           *tgbotapi.BotAPI
	allowedUserIds   []int64
	usersData        sync.Map
	stateDataManager *stateDataManager
}

func NewBotUpdateHandler(databaseHandler *database_handler.DatabaseHandler, sender bot_api.BotApi,
	botAPI *tgbotapi.BotAPI, serviceConfig *config.TalkToDBConfig) *UpdateHandler {

	result := &UpdateHandler{
		botAPI:           botAPI,
		sender:           sender,
		allowedUserIds:   serviceConfig.AllowedUserIds,
		databaseHandler:  databaseHandler,
		stateDataManager: newStateDataManager(),
	}

	return result
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
		} else if strings.HasPrefix(callback, "table-data-") {
			tableName := strings.TrimPrefix(callback, "table-data-")
			u.handleChosenTable(tableName, userID)
		} else if strings.HasPrefix(callback, "column-data-") {
			columnData := strings.Split(strings.TrimPrefix(callback, "column-data-"), "-")
			tableName := columnData[0]
			columnName := columnData[1]
			u.handleChosenColumn(tableName, columnName, userID)
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
	case "/set_description":
		u.handleSetDescriptionCommand(userID)
	default:
		u.handleStatefulMessage(text, userID)
	}

	return
}

func (u *UpdateHandler) handleStatefulMessage(text string, userID int64) {
	ok := u.handleSetDescription(text, userID)
	if ok {
		return
	}

	u.handleQuery(text, userID)
}

func (u *UpdateHandler) handleSetDescription(text string, userID int64) bool {
	defer u.stateDataManager.EmptyUserStateData(userID)

	data, ok := u.stateDataManager.GetDescriptionData(userID)
	if !ok {
		return false
	}

	err := u.databaseHandler.SetDescription(data.Table, data.Column, text)
	if err != nil {
		return false
	}
	return true
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
	defer u.stateDataManager.EmptyUserStateData(userID)
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

func createDatabaseMessageData(dbs []database_handler.Database) []messages.DatabaseData {
	var result []messages.DatabaseData
	for _, db := range dbs {
		result = append(result, messages.DatabaseData{
			Name: db.Name,
			ID:   db.ID,
		})
	}

	return result
}

func createDatabaseTablesData(db database_handler.Database) []messages.TableData {
	var result []messages.TableData
	for _, table := range db.Tables {
		result = append(result, messages.TableData{
			Name: table.Name,
		})
	}

	return result
}

func createTableColumnsData(db database_handler.Table) []messages.ColumnData {
	var result []messages.ColumnData
	for _, table := range db.Columns {
		result = append(result, messages.ColumnData{
			Name:      table.Name,
			TableName: table.Name,
		})
	}

	return result
}

func (u *UpdateHandler) handleSwitchDriver(db string, userID int64) {
	defer u.stateDataManager.EmptyUserStateData(userID)
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
	defer u.stateDataManager.EmptyUserStateData(userID)
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

func (u *UpdateHandler) Start() {
	updatesChan, err := u.botAPI.GetUpdatesChan(tgbotapi.NewUpdate(0))
	if err != nil {
		panic(err)
	}

	for update := range updatesChan {
		u.handleUpdate(&update)
	}
}

func (u *UpdateHandler) handleSetDescriptionCommand(userID int64) {
	defer u.stateDataManager.EmptyUserStateData(userID)
	database, err := u.databaseHandler.GetCurrentDatabase()
	if err != nil {
		return
	}

	u.sender.SendMessage(bot_api.Message{
		Text:        "Choose table",
		ChatId:      userID,
		ReplyMarkup: messages.GenerateDatabaseMenuButtons(createDatabaseTablesData(database)),
	})
}

func (u *UpdateHandler) handleChosenTable(tableName string, userID int64) {
	database, err := u.databaseHandler.GetCurrentDatabase()
	if err != nil {
		return
	}

	table, ok := database.GetTableByName(tableName)
	if !ok {
		return
	}

	err = u.stateDataManager.AddDescriptionTableData(tableName, userID)
	if err != nil {
		return
	}

	u.sender.SendMessage(bot_api.Message{
		Text:        fmt.Sprintf("Choose column or send description. current descrition for this table: \n%s", table.Description),
		ChatId:      userID,
		ReplyMarkup: messages.GenerateTableMenuButtons(createTableColumnsData(table)),
	})
}

func (u *UpdateHandler) handleChosenColumn(tableName string, columnName string, userID int64) {
	database, err := u.databaseHandler.GetCurrentDatabase()
	if err != nil {
		return
	}

	table, ok := database.GetTableByName(tableName)
	if !ok {
		return
	}

	_, ok = table.GetColumnByName(columnName)
	if !ok {
		return
	}

	err = u.stateDataManager.AddDescriptionColumnData(columnName, userID)
	if err != nil {
		return
	}

	u.sender.SendMessage(bot_api.Message{
		Text:   fmt.Sprintf("Send description for the selected column. current descrition for this table: \n%s", table.Description),
		ChatId: userID,
	})
}
