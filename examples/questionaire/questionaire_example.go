package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/jkevinp/tgui/button"
	"github.com/jkevinp/tgui/helper"
	"github.com/jkevinp/tgui/questionaire"
)

var qsManager *questionaire.Manager

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Load bot token from environment variable (set in .env file)
	telegramBotToken := os.Getenv("EXAMPLE_TELEGRAM_BOT_TOKEN")
	if telegramBotToken == "" {
		fmt.Println("ERROR: EXAMPLE_TELEGRAM_BOT_TOKEN is not set in .env file.")
		return
	}

	qsManager = questionaire.NewManager()

	opts := []bot.Option{
		bot.WithDefaultHandler(defaultHandler), // Handles non-command messages
		bot.WithCallbackQueryDataHandler("start_survey_cmd", bot.MatchTypeExact, startSurveyHandler),
	}

	b, err := bot.New(telegramBotToken, opts...)
	if err != nil {
		panic(err)
	}

	b.SetMyCommands(context.Background(), &bot.SetMyCommandsParams{
		Commands: []models.BotCommand{
			{Command: "survey", Description: "Start a short survey"},
		},
	})

	// Register a command handler for /survey
	b.RegisterHandler(bot.HandlerTypeMessageText, "/survey", bot.MatchTypeExact, startSurveyCommandHandler)

	// IMPORTANT: Register the questionnaire manager's message handler
	// This allows the manager to process text replies for active questionnaires.
	b.RegisterHandler(bot.HandlerTypeMessageText, "", bot.MatchTypePrefix, qsManager.HandleMessage)

	fmt.Println("Bot started! Send /survey to begin.")
	b.Start(ctx)
}

// startSurveyCommandHandler handles the /survey command
func startSurveyCommandHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}
	chatID := update.Message.Chat.ID

	// You can also present a button to start the survey
	kb := &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: "Start Survey Now", CallbackData: "start_survey_cmd"},
			},
		},
	}
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      chatID,
		Text:        "Welcome! Click the button below to start a short survey.",
		ReplyMarkup: kb,
	})
}

// startSurveyHandler is triggered by the "Start Survey Now" button
func startSurveyHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.CallbackQuery == nil || update.CallbackQuery.Message.Message == nil {
		return
	}
	chatID := update.CallbackQuery.Message.Message.Chat.ID

	// Answer the callback query to remove the "loading" state on the button
	b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: update.CallbackQuery.ID,
	})

	// Delete the message with the "Start Survey Now" button
	b.DeleteMessage(ctx, &bot.DeleteMessageParams{
		ChatID:    chatID,
		MessageID: update.CallbackQuery.Message.Message.ID,
	})

	// --- Create the Questionnaire ---
	q := questionaire.NewBuilder(chatID, qsManager).
		SetContext(ctx). // Pass context
		SetOnDoneHandler(onSurveyDone).
		SetOnCancelHandler(func() { // Inline cancel handler
			fmt.Printf("Survey cancelled for chat %d\n", chatID)
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: chatID,
				Text:   "Survey cancelled.",
			})
		}).
		// SetAllowEditAnswers(false) // Uncomment to disable editing of answered questions
		SetAllowEditAnswers(true) // Default: true (users can edit previous answers)

	// Question 1: Text input
	q.AddQuestion("name", "What is your name?", nil, validateShort)

	// Question 2: Radio buttons using ButtonGrid
	ageChoices := button.NewBuilder().
		SingleChoiceWithData("Under 18", "age_under_18").
		SingleChoiceWithData("18-30", "age_18_30").
		SingleChoiceWithData("31-45", "age_31_45").
		SingleChoiceWithData("Over 45", "age_over_45").
		Build()
	q.AddQuestion("age_group", "Which age group do you belong to?", ageChoices, nil)

	// Question 3: Checkbox using ButtonGrid (2x2 layout)
	topicChoices := button.NewBuilder().
		Row().
		ChoiceWithData("Technology", "topic_tech").
		ChoiceWithData("Sports", "topic_sports").
		Row().
		ChoiceWithData("Music", "topic_music").
		ChoiceWithData("Travel", "topic_travel").
		Build()
	q.AddMultipleAnswerQuestion("interests", "Which topics interest you? (Select multiple, then click Done)", topicChoices, nil)

	// Question 4: Yes/No choice using QuickChoices helper
	yesNoChoices := button.QuickChoices("Yes", "No")
	q.AddQuestion("developer", "Are you a software developer?", yesNoChoices, nil)

	// Question 5: Multiple choice with paired layout (2 per row)
	languageChoices := button.QuickPairedChoices("Go", "Python", "JavaScript", "Java", "C++", "Rust")
	q.AddQuestion("language", "What's your favorite programming language?", languageChoices, nil)

	// Question 6: Rating scale using custom grid layout
	ratingChoices := button.NewBuilder().
		Row().
		ChoiceWithData("⭐", "1").
		ChoiceWithData("⭐⭐", "2").
		ChoiceWithData("⭐⭐⭐", "3").
		Row().
		ChoiceWithData("⭐⭐⭐⭐", "4").
		ChoiceWithData("⭐⭐⭐⭐⭐", "5").
		Build()
	q.AddQuestion("rating", "How would you rate this questionnaire experience?", ratingChoices, nil)

	// Question 7: Text input with number validation
	q.AddQuestion("experience_years", "How many years of experience do you have with Telegram bots?", nil, validateNumber)

	// Start the questionnaire
	q.Show(ctx, b, chatID)
}

// onSurveyDone is called when all questions are answered
func onSurveyDone(ctx context.Context, b *bot.Bot, chatIDany any, answers map[string]interface{}) error {
	chatID := chatIDany.(int64) // Assuming chatID is int64

	fmt.Printf("Survey for chat %d completed. Answers:\n", chatID)
	// Create a nicely formatted response for the user
	var resultText strings.Builder
	resultText.WriteString("🎉 *Survey Completed\\!*\n\n")
	resultText.WriteString("Here are your answers:\n\n")

	// Format each answer nicely
	if name, ok := answers["name"]; ok {
		resultText.WriteString(fmt.Sprintf("👤 *Name:* %v\n", name))
	}

	if ageGroup, ok := answers["age_group"]; ok {
		ageDisplay := strings.ReplaceAll(ageGroup.(string), "_", " ")
		ageDisplay = strings.ReplaceAll(ageDisplay, "age ", "")
		resultText.WriteString(fmt.Sprintf("🎂 *Age Group:* %s\n", ageDisplay))
	}

	if interests, ok := answers["interests"]; ok {
		resultText.WriteString("🎯 *Interests:* ")
		if interestsList, ok := interests.([]interface{}); ok {
			var displayInterests []string
			for _, interest := range interestsList {
				interestStr := interest.(string)
				interestStr = strings.ReplaceAll(interestStr, "topic_", "")
				interestStr = strings.Title(interestStr)
				displayInterests = append(displayInterests, interestStr)
			}
			resultText.WriteString(strings.Join(displayInterests, ", "))
		}
		resultText.WriteString("\n")
	}

	if developer, ok := answers["developer"]; ok {
		resultText.WriteString(fmt.Sprintf("💻 *Developer:* %v\n", developer))
	}

	if language, ok := answers["language"]; ok {
		resultText.WriteString(fmt.Sprintf("🔧 *Favorite Language:* %v\n", language))
	}

	if rating, ok := answers["rating"]; ok {
		ratingNum := rating.(string)
		var stars string
		switch ratingNum {
		case "1":
			stars = "⭐ (1 star)"
		case "2":
			stars = "⭐⭐ (2 stars)"
		case "3":
			stars = "⭐⭐⭐ (3 stars)"
		case "4":
			stars = "⭐⭐⭐⭐ (4 stars)"
		case "5":
			stars = "⭐⭐⭐⭐⭐ (5 stars)"
		default:
			stars = rating.(string) + " stars"
		}
		resultText.WriteString(fmt.Sprintf("⭐ *Rating:* %s\n", stars))
	}

	if experience, ok := answers["experience_years"]; ok {
		years := experience.(string)
		if years == "1" {
			resultText.WriteString(fmt.Sprintf("🎯 *Experience:* %s year\n", years))
		} else {
			resultText.WriteString(fmt.Sprintf("🎯 *Experience:* %s years\n", years))
		}
	}

	resultText.WriteString("\nThank you for completing our survey\\! 🙏")

	// Also log to console for debugging
	for key, value := range answers {
		fmt.Printf("  %s: %v\n", key, value)
	}

	m, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    chatID,
		Text:      helper.EscapeTelegramReserved(resultText.String()),
		ParseMode: models.ParseModeMarkdown,
	})

	if err != nil {
		fmt.Printf("Error sending survey results: %v\n", err)
		return fmt.Errorf("failed to send survey results: %v", err)
	}

	fmt.Printf("Survey results sent successfully to chat %d, message ID: %d\n", chatID, m.ID)
	return nil

}

// defaultHandler catches any message that isn't a command or handled by the questionnaire manager
func defaultHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}
	// Check if there's an active questionnaire for this chat.
	// If not, or if the message is not relevant, you can reply or ignore.
	if !qsManager.Exists(update.Message.Chat.ID) {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "I'm not sure how to respond to that. Try /survey to start.",
		})
	}
	// If a questionnaire exists, qsManager.HandleMessage (registered earlier) will take care of it.
}

// --- Validator Functions ---
func validateShort(answer string) error {
	if strings.TrimSpace(answer) == "" || len(answer) < 3 {
		return fmt.Errorf("Hmm, that doesn't look like a valid answer. Please try again.")
	}
	return nil
}

func validateNumber(answer string) error {
	if _, err := strconv.Atoi(strings.TrimSpace(answer)); err != nil {
		return fmt.Errorf("Hmm, that doesn't look like a valid number. Please try again.")
	}
	return nil
}
