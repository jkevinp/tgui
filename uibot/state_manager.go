package uibot

import (
	"sync"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type UIElement interface {
	// Render returns the UI element as a string representation.
	Show(*Context) (*models.Message, error)
}

// StateManager handles thread-safe operations for managing user states.
type StateManager struct {
	Bot        *bot.Bot
	UserStates map[int64]*UserState
	Mutex      sync.RWMutex
}

// UserState represents the current state for a specific user (chat ID).
type UserState struct {
	// CurrentElement represents the currently active UI element (e.g., menu, keyboard, questionnaire)
	CurrentElement UIElement
	// PreviousElement represents the previous UI element, used for "back" actions
	PreviousElement UIElement
	// ContextData holds additional state-specific data
	ContextData map[string]interface{}
}

// NewStateManager creates a new StateManager instance.
func NewStateManager(botInstance *bot.Bot) *StateManager {
	return &StateManager{
		Bot:        botInstance,
		UserStates: make(map[int64]*UserState),
	}
}

// NewUserState creates a new UserState instance with initialized maps.
func NewUserState() *UserState {
	return &UserState{
		ContextData: make(map[string]interface{}),
	}
}

// SetState stores a state for the given chat ID.
func (sm *StateManager) SetState(chatID int64, state *UserState) {
	sm.Mutex.Lock()
	defer sm.Mutex.Unlock()
	sm.UserStates[chatID] = state
}

// GetState retrieves the state for the given chat ID.
// Returns nil if no state exists for the chat ID.
func (sm *StateManager) GetState(chatID int64) *UserState {
	sm.Mutex.RLock()
	defer sm.Mutex.RUnlock()
	return sm.UserStates[chatID]
}

// SetCurrentElement updates the current element for a chat ID,
// moving the previous current element to PreviousElement.
func (sm *StateManager) SetCurrentElement(chatID int64, element UIElement) {
	sm.Mutex.Lock()
	defer sm.Mutex.Unlock()

	state, exists := sm.UserStates[chatID]
	if !exists {
		state = NewUserState()
		sm.UserStates[chatID] = state
	}

	state.PreviousElement = state.CurrentElement
	state.CurrentElement = element
}

// Back reverts to the previous element for a chat ID.
// Returns true if there was a previous element to revert to.
func (sm *StateManager) Back(chatID int64) bool {
	sm.Mutex.Lock()
	defer sm.Mutex.Unlock()

	state, exists := sm.UserStates[chatID]
	if !exists || state.PreviousElement == nil {
		return false
	}

	state.CurrentElement = state.PreviousElement
	state.PreviousElement = nil
	return true
}

// Cancel removes all state data for a chat ID.
func (sm *StateManager) Cancel(chatID int64) {
	sm.Mutex.Lock()
	defer sm.Mutex.Unlock()
	delete(sm.UserStates, chatID)
}

// SetContextData sets context data for the current state of a chat ID.
func (sm *StateManager) SetContextData(chatID int64, key string, value interface{}) {
	sm.Mutex.Lock()
	defer sm.Mutex.Unlock()

	state, exists := sm.UserStates[chatID]
	if !exists {
		state = NewUserState()
		sm.UserStates[chatID] = state
	}

	state.ContextData[key] = value
}

// GetContextData retrieves context data for the current state of a chat ID.
// Returns nil if the key doesn't exist.
func (sm *StateManager) GetContextData(chatID int64, key string) interface{} {
	sm.Mutex.RLock()
	defer sm.Mutex.RUnlock()

	state, exists := sm.UserStates[chatID]
	if !exists {
		return nil
	}

	return state.ContextData[key]
}
