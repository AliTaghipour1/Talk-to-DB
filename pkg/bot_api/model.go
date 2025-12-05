package bot_api

import (
	"fmt"

	tgbotapi "github.com/ghiac/bale-bot-api"
)

type FileType int8

const (
	Unknown FileType = iota
	Photo
	Video
	Gif
)

type FileBytes struct {
	Bytes []byte
	Name  string
}

type Content struct {
	FileID    *string
	FileBytes *FileBytes
}

func (c Content) getContentData() (fileId string, fileBytes tgbotapi.FileBytes) {
	if c.FileID != nil {
		fileId = *c.FileID
	}

	if c.FileBytes != nil {
		fileBytes = tgbotapi.FileBytes{
			Name:  c.FileBytes.Name,
			Bytes: c.FileBytes.Bytes,
		}
	}
	return
}

type File struct {
	Content Content
	Type    FileType
}

type Message struct {
	Text             string
	ChatId           int64
	File             *File
	ReplyMarkup      interface{}
	ReplyToMessageId int
}

func (m *Message) toChattable() tgbotapi.Chattable {
	if m.File != nil {
		fileId, fileBytes := m.File.Content.getContentData()
		switch m.File.Type {
		case Photo:
			return tgbotapi.PhotoConfig{
				BaseFile: tgbotapi.BaseFile{
					BaseChat: tgbotapi.BaseChat{
						ChatID:           m.ChatId,
						ReplyMarkup:      m.ReplyMarkup,
						ReplyToMessageID: m.ReplyToMessageId,
					},
					FileID: fileId,
					File:   fileBytes,
				},
				Caption: m.Text,
			}
		case Video:
			return tgbotapi.VideoConfig{
				BaseFile: tgbotapi.BaseFile{
					BaseChat: tgbotapi.BaseChat{
						ChatID:           m.ChatId,
						ReplyMarkup:      m.ReplyMarkup,
						ReplyToMessageID: m.ReplyToMessageId,
					},
					FileID: fileId,
					File:   fileBytes,
				},
				Caption: m.Text,
			}
		case Gif:
			return tgbotapi.AnimationConfig{
				BaseFile: tgbotapi.BaseFile{
					BaseChat: tgbotapi.BaseChat{
						ChatID:           m.ChatId,
						ReplyMarkup:      m.ReplyMarkup,
						ReplyToMessageID: m.ReplyToMessageId,
					},
					FileID:      fileId,
					File:        fileBytes,
					UseExisting: true,
					MimeType:    "video/mp4",
				},
				Caption: m.Text,
			}
		default:
			return tgbotapi.DocumentConfig{
				BaseFile: tgbotapi.BaseFile{
					BaseChat: tgbotapi.BaseChat{
						ChatID:           m.ChatId,
						ReplyMarkup:      m.ReplyMarkup,
						ReplyToMessageID: m.ReplyToMessageId,
					},
					FileID:      fileId,
					File:        fileBytes,
					UseExisting: true,
				},
				Caption: m.Text,
			}
		}
	}

	message := tgbotapi.NewMessage(m.ChatId, m.Text)
	message.ReplyMarkup = m.ReplyMarkup
	message.ReplyToMessageID = m.ReplyToMessageId
	return message
}

const tracerKeyFormat = "Sender/%s"

func getTracerKey(functionName string) string {
	return fmt.Sprintf(tracerKeyFormat, functionName)
}
