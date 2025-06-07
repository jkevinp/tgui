package uibot

import (
	"context"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type Context struct {
	context.Context
	// BotInstance is the Telegram bot instance used for sending messages and handling updates
	BotInstance *UIBot
	// ChatID is the unique identifier for the chat where the UI is being displayed
	ChatID int64
	// UserID is the unique identifier for the user interacting with the UI
	UserID int64
	// MessageID is the identifier of the message that initiated the UI context
	MessageID int

	// Data holds additional context-specific data that can be used by UI elements
	Data map[string]interface{}

	Update *models.Update
}

// UIBot embeds the Telegram bot and includes a StateManager for tracking user states.
type UIBot struct {
	*bot.Bot
	StateManager *StateManager
	Middlewares  []bot.Middleware
}

// NewBotWithState creates a new BotWithState instance with an initialized StateManager.
func NewBotWithState(botInstance *bot.Bot) *UIBot {
	return &UIBot{
		Bot:          botInstance,
		StateManager: NewStateManager(botInstance),
	}
}

func (b *UIBot) Use(middleware bot.Middleware) {
	// Add the middleware to the bot's middleware stack
	b.Middlewares = append(b.Middlewares, middleware)
}

// SendMessage extends bot.SendMessage to track the message state.
func (b *UIBot) SendElement(ctx *Context, params *bot.SendMessageParams, element UIElement) (*models.Message, error) {
	// Store the element in the state manager before sending the message
	if element != nil {
		b.StateManager.SetCurrentElement(params.ChatID.(int64), element)
	}
	return b.Bot.SendMessage(ctx, params)
}

func (b *UIBot) SendMessage(ctx *Context, params *bot.SendMessageParams) (*models.Message, error) {
	return b.Bot.SendMessage(ctx, params)
}

func (b *UIBot) DeleteMessage(ctx *Context, params *bot.DeleteMessageParams) (bool, error) {
	return b.Bot.DeleteMessage(ctx, params)
}

func (b *UIBot) DeleteMessages(ctx *Context, params *bot.DeleteMessagesParams) (bool, error) {
	return b.Bot.DeleteMessages(ctx, params)
}

// HandleBack handles the "back" action for a chat ID.
// Returns true and the previous element if successful, false and nil otherwise.
func (b *UIBot) HandleBack(chatID int64) (bool, interface{}) {
	state := b.StateManager.GetState(chatID)
	if state == nil || state.PreviousElement == nil {
		return false, nil
	}

	success := b.StateManager.Back(chatID)
	if success {
		return true, state.PreviousElement
	}
	return false, nil
}

// HandleCancel handles the "cancel" action for a chat ID.
func (b *UIBot) HandleCancel(chatID int64) {
	b.StateManager.Cancel(chatID)
}

// GetCurrentElement returns the current UI element for a chat ID.
func (b *UIBot) GetCurrentElement(chatID int64) interface{} {
	state := b.StateManager.GetState(chatID)
	if state == nil {
		return nil
	}
	return state.CurrentElement
}

// SetContextData sets context data for the current state of a chat ID.
func (b *UIBot) SetContextData(chatID int64, key string, value interface{}) {
	b.StateManager.SetContextData(chatID, key, value)
}

// GetContextData retrieves context data for the current state of a chat ID.
func (b *UIBot) GetContextData(chatID int64, key string) interface{} {
	return b.StateManager.GetContextData(chatID, key)
}

type HandlerFunc func(ctx *Context, update *models.Update)

func (b *UIBot) RegisterHanderWithMiddlewares(
	handlerType bot.HandlerType,
	prefix string,
	matchType bot.MatchType,
	handler HandlerFunc,
) string {
	// Register the handler with the bot
	return b.Bot.RegisterHandler(handlerType, prefix, matchType, func(ctx context.Context, bot *bot.Bot, update *models.Update) {
		handler(
			&Context{
				Context:     ctx,
				BotInstance: b,
				ChatID:      update.Message.Chat.ID,
				UserID:      update.Message.From.ID,
				MessageID:   update.Message.ID,
				Data:        make(map[string]interface{}),
				Update:      update,
			},
			update,
		)
	}, b.Middlewares...)

}
