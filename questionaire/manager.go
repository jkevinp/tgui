package questionaire

import (
	"context"
	"fmt"
	"sync"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

// Manager handles thread-safe operations for questionnaire conversations
type Manager struct {
	mutex         sync.RWMutex
	conversations map[int64]*Questionaire
}

// NewManager creates a new Manager instance
func NewManager() *Manager {
	return &Manager{
		conversations: make(map[int64]*Questionaire),
	}
}

// Add stores a questionnaire conversation for the given chat ID
func (m *Manager) Add(chatID int64, q *Questionaire) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.conversations[chatID] = q
}

// Remove deletes the questionnaire conversation for the given chat ID
func (m *Manager) Remove(chatID int64) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	delete(m.conversations, chatID)
}

// Get retrieves the questionnaire conversation for the given chat ID
func (m *Manager) Get(chatID int64) *Questionaire {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.conversations[chatID]
}

// Exists checks if a questionnaire conversation exists for the given chat ID
func (m *Manager) Exists(chatID int64) bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	_, exists := m.conversations[chatID]
	return exists
}

// HandleMessage processes incoming messages for active questionnaire conversations
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

	if isDone := q.Answer(update.Message.Text, b, chatID); isDone {
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
