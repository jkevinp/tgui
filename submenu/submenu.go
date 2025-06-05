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

// NewSubMenuItem creates a new SubMenuItem with the provided text, callback data, and selection function.

func NewSubMenuItem(text string, callbackData string, fun inline.OnSelect) *SubMenuItem {
	return &SubMenuItem{
		Text:            text,
		CallbackData:    callbackData,
		SubMenuOnSelect: fun,
	}
}

// NewSubMenu creates a new SubMenu with the provided text and items.
// Each item is a slice of SubMenuItem pointers, allowing for multiple rows of buttons.
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

func NewBuilder(b *bot.Bot, text string) *SubMenu {
	return &SubMenu{
		Text: text,
		Kb:   inline.New(b, inline.WithPrefix("inline")),
	}
}

// Row starts a new row in the SubMenu's inline keyboard.
// This allows you to add buttons in a new row after the previous buttons.
func (m *SubMenu) Row() *SubMenu {
	m.Kb.Row()
	return m
}

// Add adds a button to the current row of the SubMenu.
func (m *SubMenu) Add(text string, callbackData string, fun inline.OnSelect) *SubMenu {
	m.Kb.Button(text, []byte(callbackData), fun)
	return m
}

// Add SubMenuItem adds a SubMenuItem to the current row of the SubMenu.
func (m *SubMenu) AddSubMenuItem(item *SubMenuItem) *SubMenu {
	m.Kb.Button(item.Text, []byte(item.CallbackData), item.SubMenuOnSelect)
	return m
}

// AddCancel adds a cancel button to the current row of the SubMenu.
func (m *SubMenu) AddCancel() *SubMenu {
	m.Kb.Row().Button("❌", []byte("cancel"), onCancel)
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
