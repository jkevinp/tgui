package datatable

import (
	"context"
	"testing"

	"github.com/go-telegram/bot"
	"github.com/jkevinp/tgui/button"
	"github.com/jkevinp/tgui/questionaire"
	"github.com/stretchr/testify/assert"
)

// Mock bot for testing
func createMockBot() *bot.Bot {
	// Since we can't create a real bot without a valid token,
	// we'll use nil and test the builder validation
	return nil
}

func TestNewBuilder(t *testing.T) {
	// Test with nil bot
	builder := NewBuilder(nil)
	assert.Nil(t, builder, "NewBuilder should return nil for nil bot")

	// Test with valid bot (we'll skip actual bot creation for unit tests)
	// In real scenarios, this would be a valid bot instance
}

func TestBuilderValidation(t *testing.T) {
	// Create a mock data handler
	mockDataHandler := func(ctx context.Context, b *bot.Bot, pageSize, pageNum int, filter map[string]interface{}) DataResult {
		return NewDataResult("test", nil, 1)
	}

	// Test missing bot (should be caught by NewBuilder)
	builder := NewBuilder(nil)
	assert.Nil(t, builder, "Builder should be nil with nil bot")

	// For the remaining tests, we'll create a builder with a placeholder bot
	// In a real test environment, you'd use a proper mock or test bot
	mockBot := &bot.Bot{} // This is a simplified mock
	builder = &DataTableBuilder{
		bot:          mockBot,
		itemsPerPage: 10,
		onError:      defaultOnError,
	}

	// Test missing data handler
	dt, err := builder.Build()
	assert.Error(t, err)
	assert.Nil(t, dt)
	assert.Contains(t, err.Error(), "DataHandler is required")

	// Test invalid items per page
	builder.itemsPerPage = 0
	builder.dataHandler = mockDataHandler
	dt, err = builder.Build()
	assert.Error(t, err)
	assert.Nil(t, dt)
	assert.Contains(t, err.Error(), "ItemsPerPage must be positive")

	// Test successful build
	builder.itemsPerPage = 10
	dt, err = builder.Build()
	assert.NoError(t, err)
	assert.NotNil(t, dt)
	assert.Equal(t, mockBot, dt.b)
	assert.Equal(t, int64(10), dt.currentFilter["pageSize"])
	assert.Equal(t, int64(1), dt.currentFilter["pageNum"])
}

func TestBuilderMethods(t *testing.T) {
	mockBot := &bot.Bot{}
	builder := &DataTableBuilder{
		bot:          mockBot,
		itemsPerPage: 10,
		onError:      defaultOnError,
	}

	mockDataHandler := func(ctx context.Context, b *bot.Bot, pageSize, pageNum int, filter map[string]interface{}) DataResult {
		return NewDataResult("test", nil, 1)
	}

	// Test WithItemsPerPage
	result := builder.WithItemsPerPage(20)
	assert.Equal(t, builder, result, "WithItemsPerPage should return the builder")
	assert.Equal(t, 20, builder.itemsPerPage)

	// Test WithItemsPerPage with invalid value (should be ignored)
	builder.WithItemsPerPage(0)
	assert.Equal(t, 20, builder.itemsPerPage, "Invalid itemsPerPage should be ignored")

	// Test WithDataHandler
	result = builder.WithDataHandler(mockDataHandler)
	assert.Equal(t, builder, result, "WithDataHandler should return the builder")
	assert.NotNil(t, builder.dataHandler)

	// Test WithFiltering
	manager := questionaire.NewManager()
	filterKeys := []string{"name", "status"}
	result = builder.WithFiltering(manager, filterKeys)
	assert.Equal(t, builder, result, "WithFiltering should return the builder")
	assert.Equal(t, manager, builder.questionaireManager)
	assert.Equal(t, filterKeys, builder.filterKeys)

	// Test WithOnErrorHandler
	customErrorHandler := func(err error) {}
	result = builder.WithOnErrorHandler(customErrorHandler)
	assert.Equal(t, builder, result, "WithOnErrorHandler should return the builder")
	assert.NotNil(t, builder.onError)

	// Test WithOnErrorHandler with nil (should be ignored)
	builder.WithOnErrorHandler(nil)
	assert.NotNil(t, builder.onError, "Nil error handler should be ignored")

	// Test WithOnCancelHandler
	customCancelHandler := func() {}
	result = builder.WithOnCancelHandler(customCancelHandler)
	assert.Equal(t, builder, result, "WithOnCancelHandler should return the builder")
	assert.NotNil(t, builder.onCancelHandler)
}

func TestBuilderChaining(t *testing.T) {
	mockBot := &bot.Bot{}
	mockDataHandler := func(ctx context.Context, b *bot.Bot, pageSize, pageNum int, filter map[string]interface{}) DataResult {
		return NewDataResult("test", nil, 1)
	}

	builder := &DataTableBuilder{
		bot:     mockBot,
		onError: defaultOnError,
	}

	// Test method chaining
	result := builder.
		WithItemsPerPage(15).
		WithDataHandler(mockDataHandler).
		WithFiltering(questionaire.NewManager(), []string{"name"}).
		WithOnCancelHandler(func() {})

	assert.Equal(t, builder, result, "Method chaining should return the same builder")
	assert.Equal(t, 15, builder.itemsPerPage)
	assert.NotNil(t, builder.dataHandler)
	assert.NotNil(t, builder.questionaireManager)
	assert.NotNil(t, builder.onCancelHandler)
	assert.Equal(t, []string{"name"}, builder.filterKeys)
}

func TestDataResultHelpers(t *testing.T) {
	// Test NewDataResult
	text := "Test text"
	replyMarkup := [][]button.Button{}
	pagesCount := int64(5)

	result := NewDataResult(text, replyMarkup, pagesCount)

	assert.Equal(t, text, result.Text)
	assert.Equal(t, replyMarkup, result.ReplyMarkup)
	assert.Equal(t, pagesCount, result.PagesCount)

	// Test NewErrorDataResult
	testError := assert.AnError
	errorResult := NewErrorDataResult(testError)

	assert.Equal(t, testError.Error(), errorResult.Text)
	assert.Nil(t, errorResult.ReplyMarkup)
	assert.Equal(t, int64(0), errorResult.PagesCount)
}
