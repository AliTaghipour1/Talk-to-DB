package bot

import (
	"log"
	"slices"

	"github.com/AliTaghipour1/Talk-to_DB/internal/config"
	"github.com/AliTaghipour1/Talk-to_DB/pkg/bot_api"
	tgbotapi "github.com/ghiac/bale-bot-api"
)

type UpdateHandler struct {
	sender         bot_api.BotApi
	botAPI         *tgbotapi.BotAPI
	allowedUserIds []int64
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
	userId := int64(update.Message.From.ID)

	if !u.isUserAllowedToUseBot(userId) {
		return
	}

	u.sender.SendMessage(bot_api.Message{
		Text:   update.Message.Text,
		ChatId: userId,
	})

	return
}

func (u *UpdateHandler) isUserAllowedToUseBot(userId int64) bool {
	return slices.Contains(u.allowedUserIds, userId)
}

func NewBotUpdateHandler(sender bot_api.BotApi, botAPI *tgbotapi.BotAPI, config *config.TalkToDBConfig) *UpdateHandler {
	result := &UpdateHandler{
		botAPI:         botAPI,
		sender:         sender,
		allowedUserIds: config.AllowedUserIds,
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
