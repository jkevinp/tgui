// Package questionaire provides an interactive questionnaire system for Telegram bots.
//
// This package allows you to create multi-step questionnaires with support for text input,
// single-choice (radio button), and multiple-choice (checkbox) questions. It includes
// features like answer validation, edit functionality, and clean ButtonGrid layouts.
//
// Basic usage:
//
//	manager := questionaire.NewManager()
//	q := questionaire.NewBuilder(chatID, manager).
//		SetOnDoneHandler(handleResults).
//		AddQuestion("name", "What's your name?", nil, validateName).
//		AddQuestion("age", "Select age group:", ageChoices, nil)
//	q.Show(ctx, bot, chatID)
//
// The package integrates with the button package for creating organized choice layouts
// using the ButtonGrid builder pattern.
package questionaire

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/go-telegram/bot/models"
	"github.com/jkevinp/tgui/button" // ButtonGrid for organized choice layouts
	"github.com/jkevinp/tgui/helper"
	"github.com/jkevinp/tgui/keyboard/inline"

	"github.com/go-telegram/bot"
	"github.com/sentimensrg/ctx/mergectx"
)

// onDoneHandlerFunc defines the signature for the completion handler function.
// It is called when all questions in the questionnaire have been answered.
// The answers map contains question keys mapped to their values.
type onDoneHandlerFunc func(ctx context.Context, b *bot.Bot, chatID any, answers map[string]interface{}) error

// QuestionFormat defines the type of input expected for a question.
type QuestionFormat int

const (
	// QuestionFormatText indicates a free-text input question where users type their response.
	QuestionFormatText QuestionFormat = 0
	// QuestionFormatRadio indicates a single-choice question with radio button style selection.
	QuestionFormatRadio QuestionFormat = 1
	// QuestionFormatCheck indicates a multiple-choice question with checkbox style selection.
	QuestionFormatCheck QuestionFormat = 2
)

// UI constants for consistent button symbols and text across the questionnaire interface.
// These can be customized by modifying the constants if different symbols are preferred.
const (
	// RadioUnselected is the symbol displayed for unselected radio button options.
	RadioUnselected = "⚪"
	// RadioSelected is the symbol displayed for the selected radio button option.
	RadioSelected = "🔘"
	// CheckUnselected is the symbol displayed for unselected checkbox options.
	CheckUnselected = "☑️"
	// CheckSelected is the symbol displayed for selected checkbox options.
	CheckSelected = "✅"
	// EditButtonText is the text displayed on edit buttons for answered questions.
	EditButtonText = "◀️ Edit"
	// DoneButtonText is the text displayed on the done button for checkbox questions.
	DoneButtonText = "✅ Done"
	// CancelButtonText is the text displayed on cancel buttons.
	CancelButtonText = "❌ Cancel"
)

// Questionaire represents an interactive questionnaire session for a specific chat.
// It manages a sequence of questions, tracks user progress, and handles answer collection.
//
// Use NewBuilder() to create a new questionnaire instance rather than constructing this struct directly.
type Questionaire struct {
	questions            []*Question       // Ordered list of questions in the questionnaire
	currentQuestionIndex int               // Index of the currently active question
	onDoneHandler        onDoneHandlerFunc // Function called when all questions are completed
	onCancelHandler      func()            // Function called when questionnaire is cancelled

	callbackID string // Unique identifier for this questionnaire's callback handlers
	msgIds     []int  // Message IDs of sent questionnaire messages for cleanup

	chatID any // Telegram chat ID where this questionnaire is running

	ctx context.Context // Context for the questionnaire session

	// InitialData contains pre-filled data that will be included in final answers
	InitialData map[string]interface{}
	// manager is the Manager instance handling this questionnaire (optional)
	manager *Manager
	// allowEditAnswers controls whether answered questions can be edited (default: true)
	allowEditAnswers bool
}

// GetAnswers returns a map of question keys to their answers or selected choices.
// For text and radio questions, the value is a string.
// For checkbox questions, the value is a slice of strings ([]string).
// The returned map also includes any InitialData that was set on the questionnaire.
func (q *Questionaire) GetAnswers() map[string]interface{} {
	answers := make(map[string]interface{})
	for _, question := range q.questions {
		if question.QuestionFormat == QuestionFormatCheck {
			answers[question.Key] = question.ChoicesSelected
		} else {
			answers[question.Key] = question.Answer
		}
	}

	for key, value := range q.InitialData {
		answers[key] = value
	}

	return answers
}

// SetInitialData sets pre-filled data for the questionnaire and returns the updated instance.
// This data will be included in the final answers map alongside user responses.
// Useful for including metadata like user ID, campaign source, etc.
func (q *Questionaire) SetInitialData(data map[string]interface{}) *Questionaire {
	q.InitialData = data
	return q
}

// Question represents a single question within a questionnaire.
// It contains the question text, possible choices (for radio/checkbox questions),
// user's answer, validation function, and formatting information.
type Question struct {
	// Key is the unique identifier for this question used in the answers map
	Key string
	// Text is the question text displayed to the user
	Text string
	// Choices contains the button layout for radio/checkbox questions (created with ButtonGrid)
	Choices [][]button.Button
	// ChoicesSelected stores selected callback data for checkbox questions
	ChoicesSelected []string
	// Answer stores the user's response (text input or selected callback data)
	Answer string
	// validator is an optional function to validate user input
	validator func(answer string) error
	// QuestionFormat determines the type of question (text, radio, or checkbox)
	QuestionFormat QuestionFormat
	// MsgID stores the Telegram message ID of the question message for editing
	MsgID int
}

// SetMsgID sets the Telegram message ID for this question.
// This is used internally to track message IDs for editing and cleanup purposes.
func (q *Question) SetMsgID(msgID int) {
	q.MsgID = msgID
}

// sendAnswerSummary adds an edit button to the completed question (if editing is enabled).
// For text questions: edits the existing message to add edit button
// For radio/checkbox questions: sends a new summary message (since original was deleted)
// If allowEditAnswers is false, no edit button is added.
func (q *Questionaire) sendAnswerSummary(ctx context.Context, b *bot.Bot, questionIndex int) {
	if questionIndex < 0 || questionIndex >= len(q.questions) {
		return
	}

	question := q.questions[questionIndex]
	if question == nil {
		return
	}

	displayAnswer := question.GetDisplayAnswer()

	// Check if editing is allowed
	var editKB *inline.Keyboard
	if q.allowEditAnswers {
		editKB = inline.New(b, inline.WithPrefix(
			fmt.Sprintf("qs_%s_answer_%d", q.callbackID, questionIndex),
		)).Button(helper.EscapeTelegramReserved(EditButtonText), []byte(fmt.Sprintf("%d", questionIndex)), q.onBack)

		if question.QuestionFormat == QuestionFormatText && question.MsgID != 0 {
			// For text questions, edit the existing message to add edit button (if enabled)
			answerText := fmt.Sprintf("✅ *%s*\n%s",
				helper.EscapeTelegramReserved(question.Text),
				helper.EscapeTelegramReserved(displayAnswer))

			_, err := b.EditMessageText(ctx, &bot.EditMessageTextParams{
				ChatID:      q.chatID,
				MessageID:   question.MsgID,
				Text:        answerText,
				ParseMode:   models.ParseModeMarkdown,
				ReplyMarkup: editKB,
			})

			if err != nil {
				// If edit fails, fall back to sending new message
				q.sendNewAnswerSummary(ctx, b, question, displayAnswer, editKB)
			}
		} else {
			// For radio/checkbox questions, send a new summary message
			q.sendNewAnswerSummary(ctx, b, question, displayAnswer, editKB)
		}
	}
}

/*
sendNewAnswerSummary sends a new message with answer summary and edit button (if provided).
Used for radio/checkbox questions or as fallback when editing fails.
editKB can be nil when editing is disabled.
*/
func (q *Questionaire) sendNewAnswerSummary(ctx context.Context, b *bot.Bot, question *Question, displayAnswer string, editKB *inline.Keyboard) {
	answerText := fmt.Sprintf("✅ *%s*\n%s",
		helper.EscapeTelegramReserved(question.Text),
		helper.EscapeTelegramReserved(displayAnswer))

	params := &bot.SendMessageParams{
		ChatID:    q.chatID,
		Text:      answerText,
		ParseMode: models.ParseModeMarkdown,
	}

	// Only add reply markup if edit keyboard is provided (editing enabled)
	if editKB != nil {
		params.ReplyMarkup = editKB
	}

	answerMsg, err := b.SendMessage(ctx, params)

	if err != nil {
		// Could optionally log error here if logging system is available
		return
	}

	q.msgIds = append(q.msgIds, answerMsg.ID)
}

// GetDisplayAnswer returns a user-friendly display string for the question's answer.
//
// The returned string is suitable for display in the Telegram chat:
//   - Text questions: Returns the user's text input or "Not answered"
//   - Radio questions: Returns the display text of the selected button or "Not selected"
//   - Checkbox questions: Returns the first selection plus count (e.g., "Tech + 2 more") or "None selected"
//
// This method converts internal callback data back to human-readable text,
// making it perfect for edit buttons and result summaries.
func (q *Question) GetDisplayAnswer() string {
	switch q.QuestionFormat {
	case QuestionFormatText:
		if q.Answer == "" {
			return "Not answered"
		}
		return q.Answer

	case QuestionFormatRadio:
		if q.Answer == "" {
			return "Not selected"
		}
		// Find the display text for the selected callback data
		for _, choiceRow := range q.Choices {
			for _, choice := range choiceRow {
				if choice.CallbackData == q.Answer {
					return choice.Text
				}
			}
		}
		return q.Answer // Fallback to callback data if not found

	case QuestionFormatCheck:
		if len(q.ChoicesSelected) == 0 {
			return "None selected"
		}
		var displayTexts []string
		// Convert callback data to display text
		for _, selected := range q.ChoicesSelected {
			for _, choiceRow := range q.Choices {
				for _, choice := range choiceRow {
					if choice.CallbackData == selected {
						displayTexts = append(displayTexts, choice.Text)
						break
					}
				}
			}
		}
		if len(displayTexts) > 0 {
			if len(displayTexts) == 1 {
				return displayTexts[0]
			}
			return fmt.Sprintf("%s \\+ %d more", displayTexts[0], len(displayTexts)-1)
		}
		return "Selected items"

	default:
		return "Unknown"
	}
}

// SetAnswer sets the answer for the question.
// This method is used internally during answer processing.
// For most use cases, answers are set automatically through user interaction.
func (q *Question) SetAnswer(answer string) {
	q.Answer = answer
}

/*
AddChoiceSelected appends a selected choice to the ChoicesSelected slice.
*/
func (q *Question) AddChoiceSelected(answer string) {
	q.ChoicesSelected = append(q.ChoicesSelected, answer)
}

func (q *Question) IsSelected(choice string) bool {
	for _, selectedChoice := range q.ChoicesSelected {
		if selectedChoice == choice {
			return true
		}
	}
	return false
}

func (q *Question) GetSelectedChoices() [][]button.Button {
	selected := make([][]button.Button, 0)

	for _, row := range q.Choices {
		newRow := make([]button.Button, 0)
		for _, choice := range row {
			if q.IsSelected(choice.CallbackData) {
				newRow = append(newRow, choice)
			}
		}
		if len(newRow) > 0 {
			selected = append(selected, newRow)
		}
	}

	return selected
}

func (q *Question) GetUnselectedChoices() [][]button.Button {
	unselected := make([][]button.Button, 0)

	for _, row := range q.Choices {
		newRow := make([]button.Button, 0)
		for _, choice := range row {
			selected := false
			for _, selectedChoice := range q.ChoicesSelected {
				if selectedChoice == choice.CallbackData {
					selected = true
					break
				}
			}
			if !selected {
				newRow = append(newRow, choice)
			}
		}
		if len(newRow) > 0 {
			unselected = append(unselected, newRow)
		}
	}

	return unselected

}

/*
Validate runs the validator function for the question, if set.
*/
func (q *Question) Validate(answer string) error {

	if q.validator != nil {

		result := q.validator(answer)
		fmt.Println("[Question] Validating answer:", answer, "for question:", q.Key, "result:", result)
		return result
	}
	return nil
}

// NewBuilder creates a new Questionaire instance for the specified chat.
//
// Parameters:
//   - chatID: The Telegram chat ID where this questionnaire will run
//   - manager: Optional Manager instance to handle text message routing (can be nil)
//
// Returns a new Questionaire instance that can be configured using the builder pattern.
//
// Example:
//
//	manager := questionaire.NewManager()
//	q := questionaire.NewBuilder(chatID, manager).
//		SetOnDoneHandler(handleCompletion).
//		AddQuestion("name", "What's your name?", nil, validateName)
func NewBuilder(chatID any, manager *Manager) *Questionaire {
	fmt.Println("new question builder:", chatID)
	return &Questionaire{
		questions:            make([]*Question, 0),
		callbackID:           "qs" + bot.RandomString(14),
		chatID:               chatID,
		currentQuestionIndex: 0,
		onDoneHandler:        nil,
		msgIds:               make([]int, 0),
		manager:              manager,
		allowEditAnswers:     true, // Default to true for backward compatibility
	}
}

// SetManager sets or updates the manager for this questionnaire and returns the updated instance.
// The manager is responsible for routing text messages to active questionnaires.
// If you're using text questions, you need a manager and must register its HandleMessage method with your bot.
func (q *Questionaire) SetManager(m *Manager) *Questionaire {
	q.manager = m
	return q
}

// SetContext sets the context for the questionnaire and returns the updated instance.
// This context will be merged with contexts passed to handler functions.
// Useful for passing request-scoped data or implementing timeouts.
func (q *Questionaire) SetContext(ctx context.Context) *Questionaire {
	q.ctx = ctx
	return q
}

// AddMultipleAnswerQuestion adds a checkbox-style question that allows multiple selections.
//
// Parameters:
//   - key: Unique identifier for this question (used in the final answers map)
//   - text: The question text shown to the user
//   - choices: Button layout created with ButtonGrid (e.g., button.NewBuilder().Row().Choice("Option1").Build())
//   - validateFunc: Optional validation function (usually nil for checkbox questions)
//
// The user can select multiple options and click "Done" to proceed.
// The answer will be stored as a []string in the final results.
//
// Example:
//
//	interests := button.NewBuilder().
//		Row().ChoiceWithData("Tech", "tech").ChoiceWithData("Sports", "sports").
//		Row().ChoiceWithData("Music", "music").ChoiceWithData("Travel", "travel").Build()
//	q.AddMultipleAnswerQuestion("interests", "Select your interests:", interests, nil)
func (q *Questionaire) AddMultipleAnswerQuestion(key string, text string, choices [][]button.Button, validateFunc func(answer string) error) *Questionaire {
	question := &Question{
		Key:             key,
		Text:            text,
		Choices:         make([][]button.Button, 0),
		ChoicesSelected: make([]string, 0),
		validator:       validateFunc,
		QuestionFormat:  QuestionFormatCheck,
	}

	question.Choices = choices

	q.questions = append(q.questions, question)

	fmt.Println("added question:", question)

	return q
}

// AddQuestion adds a text input or single-choice (radio) question to the questionnaire.
//
// Parameters:
//   - key: Unique identifier for this question (used in the final answers map)
//   - text: The question text shown to the user
//   - choices: Button layout for radio questions (nil for text input)
//   - validateFunc: Optional validation function for user input
//
// Question type is determined automatically:
//   - If choices is nil: Creates a text input question
//   - If choices is provided: Creates a radio button question (single selection)
//
// Examples:
//
//	// Text input question
//	q.AddQuestion("name", "What's your name?", nil, validateNonEmpty)
//
//	// Radio button question
//	ageChoices := button.NewBuilder().
//		SingleChoiceWithData("Under 18", "age_under_18").
//		SingleChoiceWithData("18-30", "age_18_30").Build()
//	q.AddQuestion("age_group", "Select your age group:", ageChoices, nil)
func (q *Questionaire) AddQuestion(key string, text string, choices [][]button.Button, validateFunc func(answer string) error) *Questionaire {
	question := &Question{
		Key:             key,
		Text:            text,
		Choices:         choices,
		ChoicesSelected: make([]string, 0),
		validator:       validateFunc,
	}

	if question.Choices == nil {
		question.QuestionFormat = QuestionFormatText
		question.Choices = make([][]button.Button, 0)
	} else {
		question.QuestionFormat = QuestionFormatRadio
	}

	q.questions = append(q.questions, question)

	fmt.Println("added question:", question)

	return q
}

// SetOnDoneHandler sets the completion handler called when all questions have been answered.
//
// The handler receives:
//   - ctx: Context (merged with questionnaire context if set)
//   - b: Bot instance
//   - chatID: Chat ID where questionnaire is running
//   - answers: Map of question keys to user answers
//
// Example:
//
//	q.SetOnDoneHandler(func(ctx context.Context, b *bot.Bot, chatID any, answers map[string]interface{}) error {
//		fmt.Printf("Survey completed: %+v\n", answers)
//		return b.SendMessage(ctx, &bot.SendMessageParams{
//			ChatID: chatID,
//			Text:   "Thank you for completing the survey!",
//		})
//	})
func (q *Questionaire) SetOnDoneHandler(handler onDoneHandlerFunc) *Questionaire {
	q.onDoneHandler = handler
	return q
}

// SetOnCancelHandler sets the cancellation handler called when the user cancels the questionnaire.
// The handler typically sends a cancellation message to the user.
// Questionnaire cleanup (message deletion) is handled automatically.
func (q *Questionaire) SetOnCancelHandler(handler func()) *Questionaire {
	q.onCancelHandler = handler
	return q
}

// SetAllowEditAnswers controls whether answered questions can be edited by the user.
// When set to false, answered questions will not show edit buttons and users cannot go back to modify their responses.
// When set to true (default), users can click edit buttons to modify previous answers.
//
// Parameters:
//   - allow: true to enable editing (default), false to disable editing of answered questions
//
// Example:
//
//	// Create a questionnaire without edit functionality
//	q := questionaire.NewBuilder(chatID, manager).
//		SetAllowEditAnswers(false).
//		AddQuestion("name", "What's your name?", nil, nil)
func (q *Questionaire) SetAllowEditAnswers(allow bool) *Questionaire {
	q.allowEditAnswers = allow
	return q
}

// Done is called internally when all questions have been answered.
// It marshals the answers to JSON and calls the onDoneHandler.
// This method is typically not called directly by user code.
func (q *Questionaire) Done(ctx context.Context, b *bot.Bot, update *models.Update) {

	if q.ctx != nil {
		ctx = mergectx.Join(ctx, q.ctx)
	}

	resultByte, err := GetResultByte(q)

	if err != nil {
		fmt.Println("[Questionaire] error getting result:", err)
		return
	}
	fmt.Println("[Questionaire] result of questionaire:", string(resultByte))

	if q.onDoneHandler == nil {
		fmt.Println("[Questionaire] no onDoneHandler set, skipping")
		return
	}
	result := make(map[string]interface{})
	if err := json.Unmarshal(resultByte, &result); err != nil {
		fmt.Println("[Questionaire] error unmarshalling result:", err)
	}

	if err := q.onDoneHandler(ctx, b, q.chatID, result); err != nil {
		fmt.Println("[Questionaire] error calling onDoneHandler:", err)
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: q.chatID,
			Text:   err.Error(),
		})
		return
	}

	deleteParams := bot.DeleteMessagesParams{
		ChatID:     q.chatID,
		MessageIDs: q.msgIds,
	}

	b.DeleteMessages(ctx, &deleteParams)
}

const (
	QUESTION_FORMAT = "✒️[%d/%d] %s"
)

// Show starts the questionnaire by displaying the first (or current) question to the user.
//
// This method:
//   - Registers the questionnaire with the manager (if available) for text message handling
//   - Sends the current question with appropriate keyboard layout
//   - Sets up callback handlers for button interactions
//   - Handles error display if validation failed on previous attempts
//
// Parameters:
//   - ctx: Context for the operation
//   - b: Bot instance for sending messages
//   - chatID: Telegram chat ID where to display the question
//
// The method automatically determines the question type and creates the appropriate interface:
//   - Text questions: Simple message requesting text input
//   - Radio questions: Message with inline keyboard buttons
//   - Checkbox questions: Message with selectable buttons and "Done" option
//
// Example:
//
//	q := questionaire.NewBuilder(chatID, manager).
//		AddQuestion("name", "What's your name?", nil, nil).
//		SetOnDoneHandler(handleResults)
//	q.Show(ctx, bot, chatID)
func (q *Questionaire) Show(ctx context.Context, b *bot.Bot, chatID any) {
	curQuestion := q.questions[q.currentQuestionIndex]
	fmt.Println("[question] -> ", q.callbackID, "->", curQuestion)

	if q.manager != nil && q.chatID != nil {
		q.manager.Add(q.chatID.(int64), q)
	}

	params := &bot.SendMessageParams{
		ChatID:    chatID,
		Text:      fmt.Sprintf(QUESTION_FORMAT, q.currentQuestionIndex+1, len(q.questions), helper.EscapeTelegramReserved(curQuestion.Text)),
		ParseMode: models.ParseModeMarkdown,
	}

	if ctx.Value("error") != nil {
		// If there's an error in context, append it to the question text
		errorMsg := ctx.Value("error").(string)
		params.Text = fmt.Sprintf("⚠️ *%s*", helper.EscapeTelegramReserved(errorMsg)) + "\n\n" + params.Text
		// Clear the error from context to avoid showing it again
		ctx = context.WithValue(ctx, "error", nil)
	}

	inlineKB := inline.New(b, inline.WithPrefix(
		fmt.Sprintf("qs_%s_step%d", q.callbackID, q.currentQuestionIndex),
	))

	// Handle different question formats with appropriate UI
	switch curQuestion.QuestionFormat {
	case QuestionFormatRadio:
		// Radio buttons: simple selection without checkbox symbols
		for _, choiceRow := range curQuestion.Choices {
			inlineKB.Row()
			for _, choice := range choiceRow {
				// Check if this choice is selected
				isSelected := curQuestion.Answer == choice.CallbackData
				buttonText := choice.Text
				if isSelected {
					buttonText = RadioSelected + " " + choice.Text
				} else {
					buttonText = RadioUnselected + " " + choice.Text
				}
				inlineKB.Button(
					helper.EscapeTelegramReserved(buttonText),
					[]byte(choice.CallbackData),
					q.onInlineKeyboardSelect,
				)
			}
		}

	case QuestionFormatCheck:
		// Checkbox: show selected and unselected with checkbox symbols
		// Add selected choices
		for _, choiceRow := range curQuestion.GetSelectedChoices() {
			inlineKB.Row()
			for _, choice := range choiceRow {
				inlineKB.Button(
					CheckSelected+" "+helper.EscapeTelegramReserved(choice.Text),
					[]byte(choice.CallbackData),
					q.onInlineKeyboardUnSelect,
				)
			}
		}

		// Add unselected choices
		for _, choiceRow := range curQuestion.GetUnselectedChoices() {
			inlineKB.Row()
			for _, choice := range choiceRow {
				inlineKB.Button(
					CheckUnselected+" "+helper.EscapeTelegramReserved(choice.Text),
					[]byte(choice.CallbackData),
					q.onInlineKeyboardSelect,
				)
			}
		}

		// Add "Done" button for checkbox questions
		inlineKB.Row().Button(DoneButtonText, []byte("cmd_done"), q.onDoneChoosing)

	case QuestionFormatText:
		// Text input: no buttons needed, user will type response
		break
	}

	if q.onCancelHandler != nil {
		inlineKB.Row()
		inlineKB.Button(CancelButtonText, []byte("cmd_cancel"), q.onCancel)
	}

	params.ReplyMarkup = inlineKB

	m, err := b.SendMessage(ctx, params)

	if err == nil {

		curQuestion.SetMsgID(m.ID)

		// Note: Answer summary with edit button is now handled in Answer() function
		// when we actually proceed to the next question
		q.msgIds = append(q.msgIds, m.ID)
	}

}

func (q *Questionaire) GetQuestionIndex(question *Question) int {
	for i, q := range q.questions {
		if q.Key == question.Key {
			return i
		}
	}
	return -1
}

func (q *Questionaire) onDoneChoosing(ctx context.Context, b *bot.Bot, mes models.MaybeInaccessibleMessage, data []byte) {
	b.DeleteMessage(ctx, &bot.DeleteMessageParams{
		ChatID:    q.chatID,
		MessageID: mes.Message.ID,
	})

	if isDone := q.Answer(ctx, "cmd_done", b, q.chatID); isDone {
		q.Done(ctx, b, nil)
	}

}

func (q *Questionaire) onBack(ctx context.Context, b *bot.Bot, mes models.MaybeInaccessibleMessage, data []byte) {
	b.DeleteMessage(ctx, &bot.DeleteMessageParams{
		ChatID:    q.chatID,
		MessageID: mes.Message.ID,
	})

	stepStr := string(data)

	step, err := strconv.Atoi(stepStr)
	if err != nil {
		return
	}

	if step < 0 || step >= len(q.questions) {
		return
	}

	for questionIndex, question := range q.questions {
		if questionIndex > step {
			b.DeleteMessage(ctx, &bot.DeleteMessageParams{
				ChatID:    q.chatID,
				MessageID: question.MsgID,
			})
			// Clear all answers for questions after the step we're going back to
			question.Answer = ""
			question.ChoicesSelected = make([]string, 0)
			question.MsgID = 0 // Reset message ID so it gets a new one
		}
	}

	q.currentQuestionIndex = step

	q.Show(ctx, b, q.chatID)

}

func (q *Questionaire) onInlineKeyboardSelect(ctx context.Context, b *bot.Bot, mes models.MaybeInaccessibleMessage, data []byte) {
	if isDone := q.Answer(ctx, string(data), b, q.chatID); isDone {
		q.Done(ctx, b, nil)
	}
}

func (q *Questionaire) onInlineKeyboardUnSelect(ctx context.Context, b *bot.Bot, mes models.MaybeInaccessibleMessage, data []byte) {
	curQuestion := q.questions[q.currentQuestionIndex]

	if curQuestion.QuestionFormat == QuestionFormatCheck {
		for i, selectedChoice := range curQuestion.ChoicesSelected {
			if selectedChoice == string(data) {
				curQuestion.ChoicesSelected = append(curQuestion.ChoicesSelected[:i], curQuestion.ChoicesSelected[i+1:]...)
				break
			}
		}
	}

	q.Show(ctx, b, q.chatID)

}

func (q *Questionaire) onCancel(ctx context.Context, b *bot.Bot, mes models.MaybeInaccessibleMessage, data []byte) {
	if q.onCancelHandler != nil {
		deleteParams := bot.DeleteMessagesParams{
			ChatID:     q.chatID,
			MessageIDs: q.msgIds,
		}
		b.DeleteMessages(ctx, &deleteParams)
		q.onCancelHandler()
	}
}

/*
Answer processes the user's answer for the current question and advances the questionnaire if appropriate.
Returns true if all questions have been answered.
*/
func (q *Questionaire) Answer(ctx context.Context, answer string, b *bot.Bot, chatID any) bool {
	curQuestion := q.questions[q.currentQuestionIndex]
	previousQuestionIndex := q.currentQuestionIndex

	if curQuestion.QuestionFormat == QuestionFormatCheck && answer == "cmd_done" {

		// For checkbox questions, "cmd_done" means we're advancing to next question
		q.currentQuestionIndex++
		// Send answer summary for the completed checkbox question
		q.sendAnswerSummary(ctx, b, previousQuestionIndex)
	} else if curQuestion.QuestionFormat == QuestionFormatCheck && answer != "cmd_done" {
		if err := curQuestion.Validate(answer); err != nil {

			ctx = context.WithValue(ctx, "error", err.Error())

			q.Show(ctx, b, chatID)

			return false
		}

		curQuestion.AddChoiceSelected(answer)
		// For checkbox selections, we don't advance yet, so no answer summary
	} else {
		if err := curQuestion.Validate(answer); err != nil {
			ctx = context.WithValue(ctx, "error", err.Error())
			q.Show(ctx, b, chatID)
			return false
		}

		curQuestion.SetAnswer(answer)

		if curQuestion.QuestionFormat != QuestionFormatCheck {
			// For text and radio questions, we advance immediately
			q.currentQuestionIndex++
			// Send answer summary for the completed question
			q.sendAnswerSummary(ctx, b, previousQuestionIndex)
		}
	}

	if q.currentQuestionIndex < len(q.questions) {
		q.Show(ctx, b, chatID)
	}

	return q.currentQuestionIndex >= len(q.questions)
}

// GetResultByte marshals the questionnaire answers to JSON bytes.
//
// This utility function serializes the complete answers map (including InitialData)
// to JSON format, making it suitable for logging, storage, or API transmission.
//
// Returns:
//   - []byte: JSON representation of the answers map
//   - error: JSON marshaling error, if any
//
// The JSON structure matches the answers map returned by GetAnswers():
//   - Text/Radio answers: {"key": "value"}
//   - Checkbox answers: {"key": ["option1", "option2"]}
//   - Initial data: {"key": "value"}
//
// Example usage:
//
//	jsonData, err := questionaire.GetResultByte(q)
//	if err != nil {
//		log.Printf("Failed to serialize answers: %v", err)
//		return
//	}
//	log.Printf("Survey results: %s", string(jsonData))
func GetResultByte(q *Questionaire) ([]byte, error) {
	data, err := json.Marshal(q.GetAnswers())
	return data, err
}
