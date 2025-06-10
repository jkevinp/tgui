# Plan: Implementing the Builder Pattern for DataTable

This document outlines the plan to refactor the `DataTable` component in the `tgui` package to use the Builder design pattern for its initialization. This aims to simplify the creation of `DataTable` instances, especially when dealing with multiple optional parameters, and make the API more fluent and readable.

## 1. Define `DataTableBuilder` Struct

A new struct named `DataTableBuilder` will be defined in [`datatable/datatable.go`](datatable/datatable.go:1). This struct will hold all the configuration options that can be set for a `DataTable`.

```go
// datatable/datatable.go
package datatable

import (
	"context"
	"errors" // For returning errors from Build()
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/jkevinp/tgui/button"
	"github.com/jkevinp/tgui/keyboard/inline"
	"github.com/jkevinp/tgui/questionaire"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

// Existing types like OnErrorHandler, dataHandlerFunc, DataResult, etc. remain

type DataTableBuilder struct {
	bot                 *bot.Bot
	itemsPerPage        int
	dataHandler         dataHandlerFunc
	questionaireManager *questionaire.Manager
	filterKeys          []string
	onError             OnErrorHandler
	onCancelHandler     func()
	// Potentially add other existing or new configurable fields from DataTable struct
	// e.g., custom text for buttons if that's desired
}
```

## 2. Implement Builder Methods on `DataTableBuilder`

Several methods will be added to `DataTableBuilder` to allow for a fluent configuration interface. Each `With...` method will set a corresponding field on the builder and return the builder itself for chaining.

*   **`NewBuilder(b *bot.Bot) *DataTableBuilder`**:
    *   This will be the primary entry point to start building a `DataTable`.
    *   It initializes a `DataTableBuilder` instance.
    *   The `bot.Bot` instance is a mandatory parameter for `NewBuilder` as it's fundamental for the `DataTable`.
    *   It will set sensible default values for optional parameters:
        *   `itemsPerPage`: e.g., `10`
        *   `onError`: `defaultOnError` (the existing private error handler)
        *   Other fields like `questionaireManager`, `filterKeys`, `onCancelHandler` will be `nil` or empty by default.

    ```go
    // datatable/datatable.go
    func NewBuilder(b *bot.Bot) *DataTableBuilder {
        if b == nil {
            // Or panic, or handle as per project's error strategy for invalid core components
            log.Println("[datatable] [ERROR] NewBuilder: bot instance cannot be nil")
            return nil // Or a builder that will fail on Build()
        }
        return &DataTableBuilder{
            bot:            b,
            itemsPerPage:   10, // Default items per page
            onError:        defaultOnError,
            // filterKeys, questionaireManager, onCancelHandler are nil/empty by default
        }
    }
    ```

*   **`WithItemsPerPage(count int) *DataTableBuilder`**:
    *   Sets the `itemsPerPage` field.
    *   Includes basic validation (e.g., `count > 0`).
    ```go
    // datatable/datatable.go
    func (dtb *DataTableBuilder) WithItemsPerPage(count int) *DataTableBuilder {
        if count > 0 {
            dtb.itemsPerPage = count
        }
        return dtb
    }
    ```

*   **`WithDataHandler(handler dataHandlerFunc) *DataTableBuilder`**:
    *   Sets the `dataHandler` field. This is a required component for the DataTable to function.
    ```go
    // datatable/datatable.go
    func (dtb *DataTableBuilder) WithDataHandler(handler dataHandlerFunc) *DataTableBuilder {
        dtb.dataHandler = handler
        return dtb
    }
    ```

*   **`WithFiltering(manager *questionaire.Manager, keys []string) *DataTableBuilder`**:
    *   Sets `questionaireManager` and `filterKeys`.
    *   If `manager` is `nil` or `keys` is empty, filtering capabilities might be disabled or limited.
    ```go
    // datatable/datatable.go
    func (dtb *DataTableBuilder) WithFiltering(manager *questionaire.Manager, keys []string) *DataTableBuilder {
        dtb.questionaireManager = manager
        dtb.filterKeys = keys
        return dtb
    }
    ```

*   **`WithOnErrorHandler(handler OnErrorHandler) *DataTableBuilder`**:
    *   Sets a custom `onError` handler.
    ```go
    // datatable/datatable.go
    func (dtb *DataTableBuilder) WithOnErrorHandler(handler OnErrorHandler) *DataTableBuilder {
        if handler != nil {
            dtb.onError = handler
        }
        return dtb
    }
    ```

*   **`WithOnCancelHandler(handler func()) *DataTableBuilder`**:
    *   Sets a custom `onCancelHandler`.
    ```go
    // datatable/datatable.go
    func (dtb *DataTableBuilder) WithOnCancelHandler(handler func()) *DataTableBuilder {
        dtb.onCancelHandler = handler
        return dtb
    }
    ```

## 3. Implement `Build() (*DataTable, error)` Method on `DataTableBuilder`

This is the final step in the builder process. It validates the accumulated configuration and constructs the `DataTable` instance.

*   **Validation:**
    *   Checks if all *required* fields in the `DataTableBuilder` have been set.
        *   `bot` (ensured by `NewBuilder` parameter)
        *   `dataHandler` (must be set via `WithDataHandler`)
        *   `itemsPerPage` (has a default, but could be validated if a specific range is required)
    *   If validation fails, it returns `nil` and an appropriate error (e.g., `errors.New("datatable: DataHandler is required")`).

*   **Construction:**
    *   If validation passes, it creates and initializes the `DataTable` struct using the values from the `DataTableBuilder`.
    *   The core logic currently in the existing `New()` function ([`datatable/datatable.go:116`](datatable/datatable.go:116)) for setting up:
        *   `prefix` (e.g., `dt" + bot.RandomString(14)`)
        *   Default control buttons (`CtrlBack`, `CtrlNext`, `CtrlClose`, `CtrlFilter`)
        *   `currentFilter` map initialization (with `pageSize` and `pageNum`)
        *   `filterButtons` generation based on `filterKeys`
        *   Assigning other configured values from the builder to the `DataTable` instance.
    *   This logic will be moved into the `Build()` method. The existing `New()` function will be deprecated or removed.

```go
// datatable/datatable.go
func (dtb *DataTableBuilder) Build() (*DataTable, error) {
	// Validation
	if dtb.bot == nil { // Should be caught by NewBuilder, but good for robustness
		return nil, errors.New("datatable: Bot instance is required")
	}
	if dtb.dataHandler == nil {
		return nil, errors.New("datatable: DataHandler is required")
	}
	if dtb.itemsPerPage <= 0 {
		return nil, errors.New("datatable: ItemsPerPage must be positive")
	}
    // If questionaireManager is provided, filterKeys should ideally also be provided.
    if dtb.questionaireManager != nil && (dtb.filterKeys == nil || len(dtb.filterKeys) == 0) {
        // This could be a warning or an error depending on desired strictness
        // For now, let's allow it but it might mean filtering won't work as expected.
        // dtb.onError(errors.New("datatable: QuestionaireManager provided without FilterKeys"))
    }


	// Construction (incorporating logic from the old New() function)
	prefix := "dt" + bot.RandomString(14) // Ensure bot.RandomString is accessible or reimplemented

	dt := &DataTable{
		b:                   dtb.bot,
		prefix:              prefix,
		onError:             dtb.onError, // Use configured or default
		dataHandler:         dtb.dataHandler,
		questionaireManager: dtb.questionaireManager,
		filterKeys:          dtb.filterKeys,
		onCancelHandler:     dtb.onCancelHandler,
		currentFilter:       make(map[string]interface{}),
		// Initialize control buttons (can be constants or configured via builder too)
		CtrlBack:   button.Button{Text: BACK, CallbackData: "back"},
		CtrlNext:   button.Button{Text: NEXT, CallbackData: "next"},
		CtrlClose:  button.Button{Text: CLOSE, CallbackData: "close"},
		CtrlFilter: button.Button{Text: FILTER, CallbackData: "filter"},
	}

	dt.currentFilter["pageSize"] = int64(dtb.itemsPerPage)
	dt.currentFilter["pageNum"] = int64(1) // Default to page 1

	// Build filter buttons if filterKeys are provided
	if len(dt.filterKeys) > 0 {
		filterMenu := button.NewBuilder()
		for _, filterKey := range dt.filterKeys {
			// Note: The OnClick for these buttons in the original New() was p.nagivateCallback.
			// This implies that nagivateCallback needs to be a method on DataTable.
			// The callback data needs to be constructed carefully.
			filterMenu.Row().Add(button.Button{
				Text:         filterKey,
				CallbackData: dt.prefix + "filter_" + filterKey, 
				// OnClick will be handled by the inline keyboard's general callback mechanism
				// which then calls dt.nagivateCallback based on the parsed command.
			})
		}
		filterMenu.Row().Add(button.New(
			CANCEL,
			dt.prefix+"filter_cancel",
			// OnClick will be handled similarly
		))
		dt.filterButtons = filterMenu.Build()
	}
    // Ensure defaultOnError is available if dt.onError is nil
    if dt.onError == nil {
        dt.onError = defaultOnError
    }

	return dt, nil
}
```

## 4. Deprecate/Remove Old `New()` Function

Once the builder is implemented and tested, the old `New()` function ([`datatable/datatable.go:116`](datatable/datatable.go:116)) should be marked as deprecated and eventually removed in a future version to encourage usage of the new builder pattern.

## Example Usage (Post-Refactor)

```go
// Minimal DataTable
dt, err := datatable.NewBuilder(myBot).
    WithItemsPerPage(15).
    WithDataHandler(myDataHandlerFunc).
    Build()
if err != nil {
    // handle error
}
// dt.Show(...)

// DataTable with filtering
dtWithFilters, err := datatable.NewBuilder(myBot).
    WithItemsPerPage(10).
    WithDataHandler(anotherDataHandlerFunc).
    WithFiltering(myQuestionaireManager, []string{"name", "status"}).
    WithOnCancelHandler(myCancelHandler).
    Build()
if err != nil {
    // handle error
}
// dtWithFilters.Show(...)
```

This plan provides a clear path to refactoring the `DataTable` initialization, making it more robust, flexible, and developer-friendly.