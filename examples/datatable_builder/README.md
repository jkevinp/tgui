# DataTable Builder Pattern - Interactive Demo Bot

This is a fully functional Telegram bot that demonstrates the new DataTable Builder pattern API. The bot showcases real-world usage scenarios with interactive tables, filtering, pagination, and action buttons.

## 🚀 Features Demonstrated

### 1. **Product Catalog** (`/products`)
- **Data**: 12 sample products with categories, prices, stock levels
- **Filtering**: By category, status, and name
- **Pagination**: 3 items per page
- **Actions**: Clickable buttons to view product details
- **Builder Pattern**: Shows filtering configuration

### 2. **User Management** (`/users`)
- **Data**: 8 sample users with roles, statuses, join dates
- **Filtering**: By role and status
- **Pagination**: 4 items per page
- **Actions**: Clickable buttons to manage users
- **Builder Pattern**: Demonstrates user data handling

### 3. **Simple Document List** (`/simple`)
- **Data**: 10 sample documents
- **No Filtering**: Clean example without filters
- **Pagination**: 5 items per page
- **Builder Pattern**: Shows minimal configuration

## 🛠️ Setup & Usage

### Prerequisites
- Go 1.19 or later
- Valid Telegram Bot Token

### Installation
1. Clone the repository and navigate to the example:
   ```bash
   cd examples/datatable_builder
   ```

2. Install dependencies:
   ```bash
   go mod tidy
   ```

3. Set up your bot token in `examples/.env`:
   ```bash
   EXAMPLE_TELEGRAM_BOT_TOKEN=your_bot_token_here
   ```

4. Build and run the bot:
   ```bash
   go build -o datatable_demo_bot main.go
   ./datatable_demo_bot
   ```

### Bot Commands
- `/start` - Welcome message and command overview
- `/products` - Browse interactive product catalog
- `/users` - Manage users with filtering
- `/simple` - View simple document list
- `/help` - Show available commands

## 🎯 Builder Pattern Examples

### Product Catalog (With Full Features)
```go
dt, err := datatable.NewBuilder(b).
    WithItemsPerPage(3).
    WithDataHandler(productsDataHandler).
    WithFiltering(questionaireManager, []string{"category", "status", "name"}).
    WithOnCancelHandler(func() {
        // Custom close action
    }).
    Build()
```

### User Management (Role-based Filtering)
```go
dt, err := datatable.NewBuilder(b).
    WithItemsPerPage(4).
    WithDataHandler(usersDataHandler).
    WithFiltering(questionaireManager, []string{"role", "status"}).
    Build()
```

### Simple List (No Filtering)
```go
dt, err := datatable.NewBuilder(b).
    WithItemsPerPage(5).
    WithDataHandler(simpleDataHandler).
    Build()
```

## 📊 Data Handler Implementation

Each table has its own data handler that:
1. **Applies Filters**: Processes filter parameters from user input
2. **Handles Pagination**: Calculates which items to show
3. **Formats Display**: Creates readable text with emojis and formatting
4. **Creates Actions**: Builds interactive buttons for each item
5. **Returns Results**: Provides data, buttons, and pagination info

Example data handler structure:
```go
func productsDataHandler(ctx context.Context, b *bot.Bot, pageSize, pageNum int, filter map[string]interface{}) datatable.DataResult {
    // 1. Apply filters to data
    filteredProducts := applyFilters(products, filter)
    
    // 2. Calculate pagination
    pageItems := calculatePagination(filteredProducts, pageSize, pageNum)
    
    // 3. Format display text
    text := formatProductDisplay(pageItems)
    
    // 4. Create action buttons
    buttons := createProductButtons(pageItems)
    
    // 5. Return result
    return datatable.NewDataResult(text, buttons, totalPages)
}
```

## 🔍 Interactive Features

### Filtering System
- Click **🔎 Filter** button to open filter menu
- Select filter type (category, status, name, role, etc.)
- Enter filter value through questionnaire interface
- Applied filters show as removable buttons
- Click **🗑 Filter: value** to remove specific filters

### Pagination Controls
- **⬅️** / **➡️** buttons for previous/next page
- **( n )** shows current page (clickable)
- **1 ⏮️** / **n ⏭️** buttons for first/last page (when applicable)
- Smart pagination shows 5 page buttons max

### Action Buttons
- Each data item has its own action button
- Product items: "View ProductName" → Shows product details
- User items: "Manage UserName" → Shows user management options
- Demonstrates how to integrate actions with DataTable

### Error Handling
- Clear error messages for configuration issues
- Graceful handling of missing data
- User-friendly feedback for invalid operations

## 🎨 UI/UX Features

### Rich Formatting
- **Emojis**: Visual indicators for status, types, actions
- **Markdown**: Bold text, proper formatting with escaped reserved characters
- **Structure**: Clear hierarchy and spacing
- **Icons**: Role-based and status-based visual cues
- **Text Escaping**: All user data properly escaped using `helper.EscapeTelegramReserved()`

### Responsive Design
- **Smart Text Truncation**: Long button texts are shortened
- **Adaptive Pagination**: Shows appropriate page controls
- **Context-Aware Actions**: Buttons adapt to data type
- **Mobile-Friendly**: Works well on all Telegram clients

## 🧪 Testing the Demo

1. **Start the bot** and send `/start`
2. **Try each command** to see different DataTable configurations
3. **Test filtering**:
   - Use `/products` → 🔎 Filter → category → "Electronics"
   - Use `/users` → 🔎 Filter → role → "admin"
4. **Test pagination**:
   - Navigate through multiple pages
   - Try jumping to first/last page
5. **Test actions**:
   - Click "View" buttons on products
   - Click "Manage" buttons on users
6. **Test error handling**:
   - The builder validation is shown in console logs

## 📝 Key Learning Points

This demo showcases:

✅ **Builder Pattern Benefits**:
- Clear, readable configuration
- Type-safe parameter handling
- Flexible optional settings
- Built-in validation

✅ **Real-World Integration**:
- Telegram bot command handlers
- User input processing
- Error handling and recovery
- Interactive user experience

✅ **Advanced Features**:
- Dynamic filtering with questionnaires
- Action button integration
- Rich text formatting
- Responsive pagination

✅ **Best Practices**:
- Separation of data and presentation
- Modular data handlers
- Consistent error handling
- User-friendly interfaces
- Proper text escaping for Telegram markdown

The example demonstrates how the DataTable Builder pattern makes it significantly easier to create sophisticated, interactive table interfaces in Telegram bots while maintaining clean, readable code.