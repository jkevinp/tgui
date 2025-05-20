package datatable

import (
	"github.com/go-telegram/bot/models"
)

type FilterButton struct {
	Text string
}

func (f *FilterButton) buildKB(prefix string) models.InlineKeyboardButton {
	return models.InlineKeyboardButton{
		Text:         f.Text,
		CallbackData: prefix + "filter_" + f.Text,
	}
}

type Filter struct {
	Buttons []FilterButton
}

func NewFilter(filterKeys []string) Filter {
	f := Filter{}
	for _, key := range filterKeys {
		f.Buttons = append(f.Buttons, FilterButton{
			Text: key,
		})
	}
	return f
}

func (f *Filter) buildKB(prefix string) models.InlineKeyboardMarkup {
	params := models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{},
	}

	// fmt.Println("building keyboard for filter:", len(f.Buttons))

	for _, btn := range f.Buttons {
		params.InlineKeyboard = append(params.InlineKeyboard, []models.InlineKeyboardButton{
			btn.buildKB(prefix),
		})
	}
	return params
}
