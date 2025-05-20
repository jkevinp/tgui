package datatable

import (
	"github.com/go-telegram/bot/models"
)

type Button struct {
	Text         string
	CallbackData string
	URL          string
}

func (b *Button) buildKB(prefix string) models.InlineKeyboardButton {
	return models.InlineKeyboardButton{
		Text:         b.Text,
		CallbackData: prefix + b.CallbackData,
	}
}

type Node struct {
	ID       string
	Text     string
	Keyboard [][]Button
}

func (n Node) buildKB(prefix string) models.ReplyMarkup {
	if len(n.Keyboard) == 0 {
		return nil
	}

	var kb [][]models.InlineKeyboardButton

	for _, row := range n.Keyboard {
		var kbRow []models.InlineKeyboardButton
		for _, btn := range row {
			b := models.InlineKeyboardButton{
				Text: btn.Text,
			}
			if btn.URL != "" {
				b.URL = btn.URL
			} else {
				b.CallbackData = prefix + btn.CallbackData
			}
			kbRow = append(kbRow, b)
		}
		kb = append(kb, kbRow)
	}

	return models.InlineKeyboardMarkup{
		InlineKeyboard: kb,
	}
}
