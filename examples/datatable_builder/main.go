package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/jkevinp/tgui/button"
	"github.com/jkevinp/tgui/datatable"
	"github.com/jkevinp/tgui/helper"
	"github.com/jkevinp/tgui/questionaire"
	"github.com/joho/godotenv"
)

// Sample data structures
type Product struct {
	ID          int
	Name        string
	Category    string
	Price       float64
	Stock       int
	Status      string
	Description string
}

type User struct {
	ID       int
	Name     string
	Email    string
	Role     string
	Status   string
	JoinDate string
}

// Sample data
var products = []Product{
	{1, "Laptop Pro", "Electronics", 1299.99, 5, "active", "High-performance laptop"},
	{2, "Wireless Mouse", "Electronics", 29.99, 25, "active", "Ergonomic wireless mouse"},
	{3, "Mechanical Keyboard", "Electronics", 89.99, 12, "active", "RGB mechanical keyboard"},
	{4, "Coffee Mug", "Home", 15.99, 50, "active", "Ceramic coffee mug"},
	{5, "Office Chair", "Furniture", 199.99, 8, "active", "Ergonomic office chair"},
	{6, "Standing Desk", "Furniture", 399.99, 3, "active", "Adjustable standing desk"},
	{7, "Water Bottle", "Home", 12.99, 30, "discontinued", "Stainless steel bottle"},
	{8, "Notebook", "Office", 4.99, 100, "active", "Spiral notebook"},
	{9, "Pen Set", "Office", 19.99, 40, "active", "Premium pen set"},
	{10, "Monitor", "Electronics", 299.99, 7, "active", "24-inch LCD monitor"},
	{11, "Headphones", "Electronics", 79.99, 15, "active", "Noise-canceling headphones"},
	{12, "Desk Lamp", "Home", 34.99, 20, "active", "LED desk lamp"},
}

var users = []User{
	{1, "John Doe", "john@example.com", "admin", "active", "2024-01-15"},
	{2, "Jane Smith", "jane@example.com", "user", "active", "2024-02-20"},
	{3, "Bob Johnson", "bob@example.com", "user", "active", "2024-03-10"},
	{4, "Alice Brown", "alice@example.com", "moderator", "active", "2024-01-25"},
	{5, "Charlie Wilson", "charlie@example.com", "user", "inactive", "2024-04-05"},
	{6, "Diana Davis", "diana@example.com", "user", "active", "2024-05-12"},
	{7, "Frank Miller", "frank@example.com", "moderator", "active", "2024-02-28"},
	{8, "Grace Lee", "grace@example.com", "user", "suspended", "2024-03-15"},
}

// Global variables
var questionaireManager *questionaire.Manager

// Data handler for products
func productsDataHandler(ctx context.Context, b *bot.Bot, pageSize, pageNum int, filter map[string]interface{}) datatable.DataResult {
	filteredProducts := products

	// Apply category filter
	if categoryFilter, ok := filter["category"]; ok && categoryFilter != nil {
		var filtered []Product
		searchTerm := strings.ToLower(categoryFilter.(string))
		for _, product := range products {
			if searchTerm == "" || strings.Contains(strings.ToLower(product.Category), searchTerm) {
				filtered = append(filtered, product)
			}
		}
		filteredProducts = filtered
	}

	// Apply status filter
	if statusFilter, ok := filter["status"]; ok && statusFilter != nil {
		var filtered []Product
		searchTerm := strings.ToLower(statusFilter.(string))
		for _, product := range filteredProducts {
			if searchTerm == "" || strings.Contains(strings.ToLower(product.Status), searchTerm) {
				filtered = append(filtered, product)
			}
		}
		filteredProducts = filtered
	}

	// Apply name filter
	if nameFilter, ok := filter["name"]; ok && nameFilter != nil {
		var filtered []Product
		searchTerm := strings.ToLower(nameFilter.(string))
		for _, product := range filteredProducts {
			if searchTerm == "" || strings.Contains(strings.ToLower(product.Name), searchTerm) {
				filtered = append(filtered, product)
			}
		}
		filteredProducts = filtered
	}

	// Calculate pagination
	totalItems := len(filteredProducts)
	startIndex := (pageNum - 1) * pageSize
	endIndex := startIndex + pageSize
	if endIndex > totalItems {
		endIndex = totalItems
	}

	var pageItems []Product
	if startIndex < totalItems {
		pageItems = filteredProducts[startIndex:endIndex]
	}

	// Build display text
	text := "🛍️ **Product Catalog**\n\n"
	if len(pageItems) == 0 {
		text += "No products found."
	} else {
		for _, product := range pageItems {
			status := "✅"
			if product.Status == "discontinued" {
				status = "❌"
			}
			text += fmt.Sprintf("%s **%s** (%s)\n", status, (product.Name), (product.Category))
			text += fmt.Sprintf("   💰 $%.2f | 📦 %d in stock\n", product.Price, product.Stock)
			text += fmt.Sprintf("   %s\n\n", product.Description)
		}
	}

	// Create action buttons for each product
	var buttons [][]button.Button
	if len(pageItems) > 0 {
		for _, product := range pageItems {
			actionText := fmt.Sprintf("View %s", product.Name)
			if len(actionText) > 30 {
				actionText = actionText[:27] + "..."
			}

			row := []button.Button{
				{
					Text:         actionText,
					CallbackData: fmt.Sprintf("view_product_%d", product.ID),
					OnClick: func(ctx context.Context, b *bot.Bot, mes models.MaybeInaccessibleMessage, data []byte) {
						// Handle product view action
						text := fmt.Sprintf("🛍️ You selected: %s\n\nPrice: $%.2f\nStock: %d\nDescription: %s",
							product.Name, product.Price, product.Stock, product.Description)
						_, err := b.SendMessage(ctx, &bot.SendMessageParams{
							ChatID:    mes.Message.Chat.ID,
							Text:      helper.EscapeTelegramReserved(text),
							ParseMode: models.ParseModeMarkdown,
						})
						if err != nil {
							log.Println("Error sending message:", err)
						}
					},
				},
			}
			buttons = append(buttons, row)
		}
	}

	// Calculate total pages
	totalPages := int64((totalItems + pageSize - 1) / pageSize)
	if totalPages == 0 {
		totalPages = 1
	}

	return datatable.NewDataResult(text, buttons, totalPages)
}

// Data handler for users
func usersDataHandler(ctx context.Context, b *bot.Bot, pageSize, pageNum int, filter map[string]interface{}) datatable.DataResult {
	filteredUsers := users

	// Apply role filter
	if roleFilter, ok := filter["role"]; ok && roleFilter != nil {
		var filtered []User
		searchTerm := strings.ToLower(roleFilter.(string))
		for _, user := range users {
			if searchTerm == "" || strings.Contains(strings.ToLower(user.Role), searchTerm) {
				filtered = append(filtered, user)
			}
		}
		filteredUsers = filtered
	}

	// Apply status filter
	if statusFilter, ok := filter["status"]; ok && statusFilter != nil {
		var filtered []User
		searchTerm := strings.ToLower(statusFilter.(string))
		for _, user := range filteredUsers {
			if searchTerm == "" || strings.Contains(strings.ToLower(user.Status), searchTerm) {
				filtered = append(filtered, user)
			}
		}
		filteredUsers = filtered
	}

	// Calculate pagination
	totalItems := len(filteredUsers)
	startIndex := (pageNum - 1) * pageSize
	endIndex := startIndex + pageSize
	if endIndex > totalItems {
		endIndex = totalItems
	}

	var pageItems []User
	if startIndex < totalItems {
		pageItems = filteredUsers[startIndex:endIndex]
	}

	// Build display text
	text := "👥 **User Management**\n\n"
	if len(pageItems) == 0 {
		text += "No users found."
	} else {
		for _, user := range pageItems {
			status := "✅"
			switch user.Status {
			case "inactive":
				status = "⏸️"
			case "suspended":
				status = "🚫"
			}

			roleIcon := "👤"
			switch user.Role {
			case "admin":
				roleIcon = "👑"
			case "moderator":
				roleIcon = "🛡️"
			}

			text += fmt.Sprintf("%s %s **%s**\n", status, roleIcon, (user.Name))
			text += fmt.Sprintf("   📧 %s\n", (user.Email))
			text += fmt.Sprintf("   📅 Joined: %s\n\n", user.JoinDate)
		}
	}

	// Create action buttons for each user
	var buttons [][]button.Button
	if len(pageItems) > 0 {
		for _, user := range pageItems {
			actionText := fmt.Sprintf("Manage %s", user.Name)
			if len(actionText) > 30 {
				actionText = actionText[:27] + "..."
			}

			row := []button.Button{
				{
					Text:         actionText,
					CallbackData: fmt.Sprintf("manage_user_%d", user.ID),
					OnClick: func(ctx context.Context, b *bot.Bot, mes models.MaybeInaccessibleMessage, data []byte) {
						// Handle user management action
						text := fmt.Sprintf("👥 Managing user: %s\n\nEmail: %s\nRole: %s\nStatus: %s\nJoined: %s",
							user.Name, user.Email, user.Role, user.Status, user.JoinDate)
						b.SendMessage(ctx, &bot.SendMessageParams{
							ChatID:    mes.Message.Chat.ID,
							Text:      (text),
							ParseMode: models.ParseModeMarkdown,
						})
					},
				},
			}
			buttons = append(buttons, row)
		}
	}

	// Calculate total pages
	totalPages := int64((totalItems + pageSize - 1) / pageSize)
	if totalPages == 0 {
		totalPages = 1
	}

	return datatable.NewDataResult(text, buttons, totalPages)
}

// Command handlers
func startHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}

	chatID := update.Message.Chat.ID

	text := `🎉 **Welcome to DataTable Demo!**

This bot demonstrates the new Builder pattern API for DataTable components.

Available commands:
• /products - Browse product catalog (with filtering)
• /users - User management table (with filtering)
• /simple - Simple table without filters
• /help - Show this help message

*Try the commands to see interactive tables in action!*`

	text = (text)

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    chatID,
		Text:      text,
		ParseMode: models.ParseModeMarkdown,
	})
}

func productsHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}

	chatID := update.Message.Chat.ID

	// Create DataTable with filtering
	dt, err := datatable.NewBuilder(b).
		WithItemsPerPage(3).
		WithDataHandler(productsDataHandler).
		WithFiltering(questionaireManager, []string{"category", "status", "name"}).
		WithOnCancelHandler(func() {
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: chatID,
				Text:   "❌ Product catalog closed.",
			})
		}).
		Build()

	if err != nil {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   fmt.Sprintf("❌ Error creating product table: %v", err),
		})
		return
	}

	// Show the DataTable
	_, err = dt.Show(ctx, b, chatID, map[string]interface{}{
		"pageSize": 3,
		"pageNum":  1,
	})

	if err != nil {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   fmt.Sprintf("❌ Error showing product table: %v", err),
		})
	}
}

func usersHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}

	chatID := update.Message.Chat.ID

	// Create DataTable with filtering
	dt, err := datatable.NewBuilder(b).
		WithItemsPerPage(4).
		WithDataHandler(usersDataHandler).
		WithFiltering(questionaireManager, []string{"role", "status"}).
		WithOnCancelHandler(func() {
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: chatID,
				Text:   "❌ User management closed.",
			})
		}).
		Build()

	if err != nil {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   fmt.Sprintf("❌ Error creating user table: %v", err),
		})
		return
	}

	// Show the DataTable
	_, err = dt.Show(ctx, b, chatID, map[string]interface{}{
		"pageSize": 4,
		"pageNum":  1,
	})

	if err != nil {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   fmt.Sprintf("❌ Error showing user table: %v", err),
		})
	}
}

func simpleHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}

	chatID := update.Message.Chat.ID

	// Simple data handler without filters
	simpleDataHandler := func(ctx context.Context, b *bot.Bot, pageSize, pageNum int, filter map[string]interface{}) datatable.DataResult {
		items := []string{
			"📝 Document 1", "📝 Document 2", "📝 Document 3", "📝 Document 4", "📝 Document 5",
			"📝 Document 6", "📝 Document 7", "📝 Document 8", "📝 Document 9", "📝 Document 10",
		}

		// Calculate pagination
		totalItems := len(items)
		startIndex := (pageNum - 1) * pageSize
		endIndex := startIndex + pageSize
		if endIndex > totalItems {
			endIndex = totalItems
		}

		var pageItems []string
		if startIndex < totalItems {
			pageItems = items[startIndex:endIndex]
		}

		// Build display text
		text := "📄 **Simple Document List**\n\n"
		if len(pageItems) == 0 {
			text += "No documents found."
		} else {
			for i, item := range pageItems {
				text += fmt.Sprintf("%d. %s\n", startIndex+i+1, item)
			}
		}

		// Calculate total pages
		totalPages := int64((totalItems + pageSize - 1) / pageSize)
		if totalPages == 0 {
			totalPages = 1
		}

		return datatable.NewDataResult(text, nil, totalPages)
	}

	// Create simple DataTable without filtering
	dt, err := datatable.NewBuilder(b).
		WithItemsPerPage(5).
		WithDataHandler(simpleDataHandler).
		WithOnCancelHandler(func() {
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: chatID,
				Text:   "❌ Document list closed.",
			})
		}).
		Build()

	if err != nil {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   fmt.Sprintf("❌ Error creating simple table: %v", err),
		})
		return
	}

	// Show the DataTable
	_, err = dt.Show(ctx, b, chatID, map[string]interface{}{
		"pageSize": 5,
		"pageNum":  1,
	})

	if err != nil {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   fmt.Sprintf("❌ Error showing simple table: %v", err),
		})
	}
}

func helpHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	startHandler(ctx, b, update)
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Load environment variables from .env file
	envPath := filepath.Join("..", ".env")
	if err := godotenv.Load(envPath); err != nil {
		log.Printf("Warning: Could not load .env file from %s: %v", envPath, err)
		log.Println("Make sure the .env file exists in the examples directory")
	}

	// Get bot token from environment
	token := os.Getenv("EXAMPLE_TELEGRAM_BOT_TOKEN")
	if token == "" {
		log.Fatal("EXAMPLE_TELEGRAM_BOT_TOKEN environment variable is required. Please check examples/.env file.")
	}

	// Create questionaire manager for filtering
	questionaireManager = questionaire.NewManager()

	// Configure bot with handlers
	opts := []bot.Option{
		bot.WithDefaultHandler(func(ctx context.Context, b *bot.Bot, update *models.Update) {
			if update.Message == nil {
				return
			}

			// Check if there's an active questionnaire or datatable for this chat
			if !questionaireManager.Exists(update.Message.Chat.ID) {
				b.SendMessage(ctx, &bot.SendMessageParams{
					ChatID: update.Message.Chat.ID,
					Text:   "ℹ️ Use /start to see available commands, or try /products, /users, or /simple to see DataTable examples.",
				})
			}
		}),
	}

	// Create bot instance
	b, err := bot.New(token, opts...)
	if err != nil {
		log.Fatal("Failed to create bot:", err)
	}

	// Set bot commands
	b.SetMyCommands(ctx, &bot.SetMyCommandsParams{
		Commands: []models.BotCommand{
			{Command: "start", Description: "Show welcome message and available commands"},
			{Command: "products", Description: "Browse product catalog with filtering"},
			{Command: "users", Description: "User management table with filtering"},
			{Command: "simple", Description: "Simple document list without filters"},
			{Command: "help", Description: "Show help information"},
		},
	})

	// Register command handlers
	b.RegisterHandler(bot.HandlerTypeMessageText, "/start", bot.MatchTypeExact, startHandler)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/products", bot.MatchTypeExact, productsHandler)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/users", bot.MatchTypeExact, usersHandler)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/simple", bot.MatchTypeExact, simpleHandler)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/help", bot.MatchTypeExact, helpHandler)

	// Register questionnaire manager handler for text input (filtering)
	b.RegisterHandler(bot.HandlerTypeMessageText, "", bot.MatchTypePrefix, questionaireManager.HandleMessage)

	log.Println("🚀 DataTable Demo Bot started!")
	log.Println("Available commands:")
	log.Println("  /start - Welcome message")
	log.Println("  /products - Product catalog with filtering")
	log.Println("  /users - User management with filtering")
	log.Println("  /simple - Simple list without filtering")
	log.Println("  /help - Show help")
	log.Println()
	log.Println("The bot demonstrates the new DataTable Builder pattern with:")
	log.Println("  ✨ Fluent API for easy configuration")
	log.Println("  🔍 Interactive filtering capabilities")
	log.Println("  📄 Pagination controls")
	log.Println("  🎯 Action buttons on data items")
	log.Println("  🛡️ Built-in validation and error handling")

	// Start the bot
	b.Start(ctx)
}
