package submenu

import (
	"context"

	"github.com/jkevinp/tgui/keyboard/inline"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type SubMenu struct {
	Kb    *inline.Keyboard
	Text  string
	MsgID any
}

type SubMenuItem struct {
	Text            string
	CallbackData    string
	SubMenuOnSelect inline.OnSelect
}

// type SubMenuOnSelect func(ctx context.Context, b *bot.Bot, update *models.Update, data []byte)

func NewSubMenuItem(text string, callbackData string, fun inline.OnSelect) *SubMenuItem {
	return &SubMenuItem{
		Text:            text,
		CallbackData:    callbackData,
		SubMenuOnSelect: fun,
	}
}

func NewSubMenu(b *bot.Bot, text string, items ...[]*SubMenuItem) *SubMenu {
	m := &SubMenu{
		Text: text,
	}

	inlineKB := inline.New(b, inline.WithPrefix("inline"))

	for _, item := range items {
		inlineKB.Row()
		for _, i := range item {
			inlineKB.Button(i.Text, []byte(i.CallbackData), i.SubMenuOnSelect)
		}
	}

	inlineKB.Row().Button("❌", []byte("cancel"), onCancel)

	m.Kb = inlineKB

	return m

}

func onCancel(ctx context.Context, b *bot.Bot, mes models.MaybeInaccessibleMessage, data []byte) {
	b.DeleteMessage(ctx, &bot.DeleteMessageParams{
		ChatID:    mes.Message.Chat.ID,
		MessageID: mes.Message.ID,
	})
}

func (m *SubMenu) Show(ctx context.Context, b *bot.Bot, chatID any) (*models.Message, error) {
	msg, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      chatID,
		Text:        m.Text,
		ReplyMarkup: m.Kb,
	})

	if err == nil {
		m.MsgID = msg.ID
	}

	return msg, err
}
