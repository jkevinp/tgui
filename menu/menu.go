package menu

import (
	"context"

	"github.com/jkevinp/tgui/keyboard/reply"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type Menu struct {
	Kb          *reply.ReplyKeyboard
	Text        string
	botInstance *bot.Bot
}

// OnMenuItemSelect defines a function type for handling menu item selection.
// It takes a context, a bot instance, an update containing the selected item,
// and the selected MenuItem. It returns an error if any occurs during processing.
// This function is typically used to define custom behavior when a user selects

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
		Text:        text,
		botInstance: b,
	}

	demoReplyKeyboard := reply.New(
		reply.WithPrefix("reply_keyboard"),
		reply.IsSelective(),
		reply.IsPersistent(),
		reply.ResizableKeyboard(),
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

func NewBuilder(b *bot.Bot, text string) *Menu {
	return &Menu{
		Text:        text,
		botInstance: b,
		Kb: reply.New(
			reply.WithPrefix("menu"),
			reply.IsSelective(),
			reply.IsPersistent(),
			reply.ResizableKeyboard(),
		),
	}
}

func (m *Menu) Row() *Menu {
	m.Kb.Row()
	return m
}
func (m *Menu) Add(text string, handler bot.HandlerFunc) *Menu {

	m.Kb.Button(text, m.botInstance, bot.MatchTypeExact, handler)

	if handler != nil {
		m.botInstance.RegisterHandler(
			bot.HandlerTypeMessageText,
			text,
			bot.MatchTypeExact,
			handler,
		)
	}
	return m
}
