package messages

import (
	"fmt"

	tgbotapi "github.com/ghiac/bale-bot-api"
)

type DatabaseData struct {
	Name string
	ID   int
}

func (d DatabaseData) button() tgbotapi.InlineKeyboardButton {
	return createButton(d.Name, fmt.Sprintf("database-data-%d", d.ID))
}

func GenerateDatabaseButtons(dbs []DatabaseData) tgbotapi.InlineKeyboardMarkup {
	var result [][]tgbotapi.InlineKeyboardButton
	for _, db := range dbs {
		result = append(result, []tgbotapi.InlineKeyboardButton{db.button()})
	}

	return tgbotapi.InlineKeyboardMarkup{InlineKeyboard: result}
}

type TableData struct {
	Name string
}

func (d TableData) button() tgbotapi.InlineKeyboardButton {
	return createButton(d.Name, fmt.Sprintf("table-data-%s", d.Name))
}

func GenerateDatabaseMenuButtons(dbs []TableData) tgbotapi.InlineKeyboardMarkup {
	var result [][]tgbotapi.InlineKeyboardButton
	for _, db := range dbs {
		result = append(result, []tgbotapi.InlineKeyboardButton{db.button()})
	}

	return tgbotapi.InlineKeyboardMarkup{InlineKeyboard: result}
}

type ColumnData struct {
	Name      string
	TableName string
}

func (d ColumnData) button() tgbotapi.InlineKeyboardButton {
	return createButton(d.Name, fmt.Sprintf("column-data-%s-%s", d.Name, d.TableName))
}

func GenerateTableMenuButtons(dbs []ColumnData) tgbotapi.InlineKeyboardMarkup {
	var result [][]tgbotapi.InlineKeyboardButton
	for _, db := range dbs {
		result = append(result, []tgbotapi.InlineKeyboardButton{db.button()})
	}

	return tgbotapi.InlineKeyboardMarkup{InlineKeyboard: result}
}

func createStaticButton(text string) tgbotapi.InlineKeyboardButton {
	return tgbotapi.InlineKeyboardButton{
		Text:         text,
		CallbackData: &text,
	}
}

func createButton(text string, callback string) tgbotapi.InlineKeyboardButton {
	return tgbotapi.InlineKeyboardButton{
		Text:         text,
		CallbackData: &callback,
	}
}
