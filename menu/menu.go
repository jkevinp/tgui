package menu

import (
	"context"

	"github.com/jkevinp/tgui/keyboard/reply"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type Menu struct {
	Kb   *reply.ReplyKeyboard
	Text string
}

type MenuItem struct {
	Text string
}

func NewMenuItem(text string) *MenuItem {
	return &MenuItem{
		Text: text,
	}
}

func NewMenu(b *bot.Bot, text string, items ...[]*MenuItem) *Menu {
	m := &Menu{
		Text: text,
	}

	demoReplyKeyboard := reply.New(
		reply.WithPrefix("reply_keyboard"),
		reply.IsSelective(),
		reply.IsOneTimeKeyboard(),
	)

	for _, item := range items {
		demoReplyKeyboard.Row()
		for _, i := range item {
			demoReplyKeyboard.Button(i.Text, b, bot.MatchTypeExact, nil)
		}
	}

	m.Kb = demoReplyKeyboard

	return m

}

func (m *Menu) Show(ctx context.Context, b *bot.Bot, chatID any) (*models.Message, error) {
	return b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      chatID,
		Text:        m.Text,
		ReplyMarkup: m.Kb,
	})
}

// func onReplyKeyboardSelect(ctx context.Context, b *bot.Bot, update *models.Update) {
// 	b.SendMessage(ctx, &bot.SendMessageParams{
// 		ChatID: update.Message.Chat.ID,
// 		Text:   "You selected: " + string(update.Message.Text),
// 	})
// }
