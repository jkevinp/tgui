# DataTable Component

The DataTable component provides an interactive, paginated table interface for Telegram bots with built-in filtering capabilities.

## ✨ New Builder Pattern API

We've introduced a new Builder pattern API that makes creating DataTable instances more flexible and readable. The old `New()` function is still available but deprecated.

### Basic Usage

```go
// Simple DataTable without filtering
dt, err := datatable.NewBuilder(bot).
    WithItemsPerPage(10).
    WithDataHandler(myDataHandler).
    Build()
if err != nil {
    log.Fatal("Failed to build DataTable:", err)
}

// Show the DataTable
message, err := dt.Show(ctx, bot, chatID, map[string]interface{}{
    "pageSize": 10,
    "pageNum":  1,
})
```

### Advanced Usage with Filtering

```go
// DataTable with filtering capabilities
dt, err := datatable.NewBuilder(bot).
    WithItemsPerPage(5).
    WithDataHandler(myDataHandler).
    WithFiltering(questionaireManager, []string{"name", "status", "date"}).
    WithOnCancelHandler(func() {
        log.Println("User cancelled the DataTable")
    }).
    WithOnErrorHandler(func(err error) {
        log.Printf("DataTable error: %v", err)
    }).
    Build()
```

## Builder Methods

### Required Methods

- **`NewBuilder(bot *bot.Bot)`** - Creates a new DataTableBuilder with the bot instance
- **`WithDataHandler(handler dataHandlerFunc)`** - Sets the data fetching function (required)
- **`Build()`** - Validates configuration and creates the DataTable instance

### Optional Methods

- **`WithItemsPerPage(count int)`** - Sets items per page (default: 10)
- **`WithFiltering(manager *questionaire.Manager, keys []string)`** - Enables filtering
- **`WithOnErrorHandler(handler OnErrorHandler)`** - Sets custom error handler
- **`WithOnCancelHandler(handler func())`** - Sets custom cancel handler

## Data Handler Function

The data handler function is responsible for fetching and formatting your data:

```go
func myDataHandler(ctx context.Context, b *bot.Bot, pageSize, pageNum int, filter map[string]interface{}) datatable.DataResult {
    // Fetch your data based on pagination and filters
    items := fetchItems(pageSize, pageNum, filter)
    
    // Format the display text
    text := formatItemsAsText(items)
    
    // Create buttons for each item (optional)
    buttons := createItemButtons(items)
    
    // Calculate total pages
    totalPages := calculateTotalPages(totalItemCount, pageSize)
    
    return datatable.NewDataResult(text, buttons, totalPages)
}
```

## Benefits of the Builder Pattern

### ✅ **More Readable**
```go
// Old way (deprecated)
dt := datatable.New(bot, 10, handler, manager, []string{"name"})

// New way (recommended)
dt, err := datatable.NewBuilder(bot).
    WithItemsPerPage(10).
    WithDataHandler(handler).  
    WithFiltering(manager, []string{"name"}).
    Build()
```

### 🛡️ **Built-in Validation**
The `Build()` method validates that all required parameters are provided:
```go
dt, err := datatable.NewBuilder(bot).
    WithItemsPerPage(10).
    // Missing WithDataHandler - Build() will return an error
    Build()
if err != nil {
    log.Printf("Error: %v", err) // "datatable: DataHandler is required"
}
```

### 🔧 **Flexible Configuration**
Only specify the options you need:
```go
// Minimal configuration
dt, _ := datatable.NewBuilder(bot).WithDataHandler(handler).Build()

// Full configuration
dt, _ := datatable.NewBuilder(bot).
    WithItemsPerPage(20).
    WithDataHandler(handler).
    WithFiltering(manager, []string{"name", "status"}).
    WithOnErrorHandler(customErrorHandler).
    WithOnCancelHandler(customCancelHandler).
    Build()
```

## Migration from Old API

If you're using the old `New()` function, migration is straightforward:

```go
// Before
dt := datatable.New(bot, 10, handler, manager, []string{"name", "date"})

// After  
dt, err := datatable.NewBuilder(bot).
    WithItemsPerPage(10).
    WithDataHandler(handler).
    WithFiltering(manager, []string{"name", "date"}).
    Build()
if err != nil {
    // Handle error
}
```

## Interactive Demo

🎯 **Try the Interactive Demo Bot!**

We've created a fully functional Telegram bot that demonstrates all DataTable Builder pattern features in real scenarios:

- **Product Catalog** with filtering by category, status, and name
- **User Management** with role and status filtering
- **Simple Document List** without filtering
- **Action Buttons** on each data item
- **Real-time Filtering** through questionnaire interface
- **Rich Formatting** with emojis and markdown

### Quick Start
```bash
cd examples/datatable_builder
go build -o datatable_demo_bot main.go
./datatable_demo_bot
```

**Commands to try:**
- `/start` - Welcome and overview
- `/products` - Interactive product catalog
- `/users` - User management table
- `/simple` - Simple list example

See [`examples/datatable_builder/README.md`](../examples/datatable_builder/README.md) for detailed setup instructions and feature explanations.

### Code Examples
The interactive demo showcases real implementations:
- Complex data handlers with filtering logic
- Action button integration
- Error handling and user feedback
- Multiple DataTable configurations in one bot

This demonstrates how the Builder pattern makes creating sophisticated table interfaces significantly easier compared to the old API.

## Error Handling

The Builder pattern provides clear error messages for common mistakes:

- `"datatable: Bot instance is required"` - NewBuilder() called with nil bot
- `"datatable: DataHandler is required"` - Build() called without setting data handler
- `"datatable: ItemsPerPage must be positive"` - Invalid items per page value

## Backward Compatibility

The old `New()` function is still available for backward compatibility but is marked as deprecated. It will be removed in a future version, so please migrate to the Builder pattern.