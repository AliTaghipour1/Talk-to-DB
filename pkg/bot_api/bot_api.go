package bot_api

import (
	"WordSearchBot/pkg/tracer"
	"log"
	"time"

	tgbotapi "github.com/ghiac/bale-bot-api"
)

type BotApi interface {
	SendMessage(message Message) (int, error)
	EditMessage(text string, chatID int64, messageID int, replyMarkup tgbotapi.InlineKeyboardMarkup) error
	SendCallbackAlert(callbackQueryId string, text string) error
	IsMember(userId int64, channelID int64) (bool, error)
}

type SenderBaleBotImpl struct {
	bot    *tgbotapi.BotAPI
	tracer *tracer.Tracer
}

func (s *SenderBaleBotImpl) IsMember(userId int64, channelId int64) (bool, error) {
	_, err := s.bot.GetChatMember(tgbotapi.ChatConfigWithUser{
		ChatID: channelId,
		UserID: int(userId),
	})

	if err != nil {
		if err.Error() == "Bad Request: no such group or user" {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func (s *SenderBaleBotImpl) SendCallbackAlert(callbackQueryId string, text string) error {
	t := time.Now()
	defer func() {
		s.tracer.SendEvent(&tracer.Event{
			Key:   getTracerKey("SendCallbackAlert"),
			Value: time.Since(t),
		})
	}()
	alert := tgbotapi.NewCallbackWithAlert(callbackQueryId, text)

	_, err := s.bot.AnswerCallbackQuery(alert)
	if err != nil {
		log.Println("error sending callback answer: ", err)
		return err
	}
	return nil
}

func (s *SenderBaleBotImpl) EditMessage(text string, chatID int64, messageID int, replyMarkup tgbotapi.InlineKeyboardMarkup) error {
	t := time.Now()
	defer func() {
		s.tracer.SendEvent(&tracer.Event{
			Key:   getTracerKey("EditMessage"),
			Value: time.Since(t),
		})
	}()

	message := tgbotapi.EditMessageTextConfig{
		BaseEdit: tgbotapi.BaseEdit{
			ChatID:      chatID,
			MessageID:   messageID,
			ReplyMarkup: &replyMarkup,
		},
		Text: text,
	}

	_, err := s.bot.Send(message)
	if err != nil {
		log.Println("error sending edit: ", err)
		return err
	}
	return nil
}

func (s *SenderBaleBotImpl) SendMessage(message Message) (int, error) {
	t := time.Now()
	defer func() {
		s.tracer.SendEvent(&tracer.Event{
			Key:   getTracerKey("SendMessage"),
			Value: time.Since(t),
		})
	}()
	send, err := s.bot.Send(message.toChattable())
	if err != nil {
		log.Println("error sending message: ", err)
		return 0, err
	}

	return send.MessageID, nil
}

func NewSenderBot(bot *tgbotapi.BotAPI) BotApi {
	return &SenderBaleBotImpl{
		bot:    bot,
		tracer: tracer.GetTracer(),
	}
}
