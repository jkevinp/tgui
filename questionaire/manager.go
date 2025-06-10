package questionaire

import (
	"context"
	"fmt"
	"sync"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

// Manager handles thread-safe operations for questionnaire conversations.
// It manages active questionnaire sessions across multiple chats and routes
// incoming text messages to the appropriate questionnaire instance.
//
// A Manager is required for questionnaires that include text input questions,
// as it handles the routing of user text replies to the correct questionnaire.
type Manager struct {
	mutex         sync.RWMutex            // Protects concurrent access to conversations map
	conversations map[int64]*Questionaire // Maps chat IDs to active questionnaire instances
}

// NewManager creates a new Manager instance for handling questionnaire sessions.
// Typically, you create one Manager when your bot starts and register its
// HandleMessage method with your bot to process text messages.
//
// Example:
//
//	manager := questionaire.NewManager()
//	bot.RegisterHandler(bot.HandlerTypeMessageText, "", bot.MatchTypePrefix, manager.HandleMessage)
func NewManager() *Manager {
	return &Manager{
		conversations: make(map[int64]*Questionaire),
	}
}

// Add stores a questionnaire conversation for the given chat ID.
// This is called automatically when a questionnaire with a manager is shown.
// Multiple questionnaires can be active simultaneously in different chats.
func (m *Manager) Add(chatID int64, q *Questionaire) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.conversations[chatID] = q
}

// Remove deletes the questionnaire conversation for the given chat ID.
// This is called automatically when a questionnaire completes or is cancelled.
// After removal, text messages from this chat will no longer be routed to a questionnaire.
func (m *Manager) Remove(chatID int64) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	delete(m.conversations, chatID)
}

// Get retrieves the questionnaire conversation for the given chat ID.
// Returns nil if no active questionnaire exists for the chat.
// This method is thread-safe and can be called concurrently.
func (m *Manager) Get(chatID int64) *Questionaire {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.conversations[chatID]
}

// Exists checks if a questionnaire conversation exists for the given chat ID.
// Returns true if an active questionnaire is running in the specified chat.
// This method is thread-safe and can be called concurrently.
func (m *Manager) Exists(chatID int64) bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	_, exists := m.conversations[chatID]
	return exists
}

// HandleMessage processes incoming text messages for active questionnaire conversations.
// This method should be registered with your bot to handle text message updates:
//
//	bot.RegisterHandler(bot.HandlerTypeMessageText, "", bot.MatchTypePrefix, manager.HandleMessage)
//
// The method:
//   - Checks if an active questionnaire exists for the message's chat ID
//   - Routes the message text to the appropriate questionnaire's Answer method
//   - Handles questionnaire completion and cleanup automatically
//   - Ignores messages from chats without active questionnaires
//
// This enables seamless text input handling for questionnaire text questions
// without requiring manual message routing in your application code.
func (m *Manager) HandleMessage(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}

	chatID := update.Message.Chat.ID

	q, exists := m.conversations[chatID]
	if !exists {
		return
	}

	fmt.Printf("[questionaire manager] ChatID: %v, Message: %v, Active conversations: %d\n",
		chatID, update.Message.Text, len(m.conversations))

	if isDone := q.Answer(ctx, update.Message.Text, b, chatID); isDone {
		result, err := GetResultByte(q)
		if err != nil {
			fmt.Println("[questionaire manager] error getting result:", err)
			return
		}
		fmt.Println("[questionaire manager] result of questionaire:", string(result))

		q.Done(ctx, b, update)

		q.manager.Remove(chatID)

		fmt.Printf("[questionaire manager] session %v deleted, remaining sessions: %d\n",
			chatID, len(m.conversations))
	}
}
