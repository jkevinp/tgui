package questionaire

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/go-telegram/bot/models"
	"github.com/jkevinp/tgui/button"
	"github.com/jkevinp/tgui/helper"
	"github.com/jkevinp/tgui/keyboard/inline"

	"github.com/go-telegram/bot"
	"github.com/sentimensrg/ctx/mergectx"
)

type onDoneHandlerFunc func(ctx context.Context, b *bot.Bot, chatID any, answers map[string]interface{}) error

type QuestionFormat int

const (
	QuestionFormatText  QuestionFormat = 0
	QuestionFormatRadio QuestionFormat = 1
	QuestionFormatCheck QuestionFormat = 2
)

// UI constants for better maintainability
const (
	RadioUnselected  = "⚪"
	RadioSelected    = "🔘"
	CheckUnselected  = "☑️"
	CheckSelected    = "✅"
	EditButtonText   = "◀️ Edit"
	DoneButtonText   = "✅ Done"
	CancelButtonText = "❌ Cancel"
)

type Questionaire struct {
	questions            []*Question
	currentQuestionIndex int
	onDoneHandler        onDoneHandlerFunc
	onCancelHandler      func()

	callbackID string
	msgIds     []int

	chatID any

	ctx context.Context

	InitialData map[string]interface{}
	manager     *Manager
}

/*
GetAnswers returns a map of question keys to their answers or selected choices.
*/
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

/*
SetInitialData sets the initial data for the questionnaire and returns the updated instance.
*/
func (q *Questionaire) SetInitialData(data map[string]interface{}) *Questionaire {
	q.InitialData = data
	return q
}

type Question struct {
	Key             string
	Text            string
	Choices         [][]button.Button
	ChoicesSelected []string
	Answer          string
	validator       func(answer string) error
	QuestionFormat  QuestionFormat

	MsgID int // Message ID of the question message, if applicable
}

func (q *Question) SetMsgID(msgID int) {
	q.MsgID = msgID
}

/*
GetDisplayAnswer returns a user-friendly display string for the question's answer.
*/
/*
sendAnswerSummary adds an edit button to the completed question.
For text questions: edits the existing message to add edit button
For radio/checkbox questions: sends a new summary message (since original was deleted)
*/
func (q *Questionaire) sendAnswerSummary(ctx context.Context, b *bot.Bot, questionIndex int) {
	if questionIndex < 0 || questionIndex >= len(q.questions) {
		return
	}

	question := q.questions[questionIndex]
	if question == nil {
		return
	}

	displayAnswer := question.GetDisplayAnswer()
	editKB := inline.New(b, inline.WithPrefix(
		fmt.Sprintf("qs_%s_answer_%d", q.callbackID, questionIndex),
	)).Button(helper.EscapeTelegramReserved(EditButtonText), []byte(fmt.Sprintf("%d", questionIndex)), q.onBack)

	if question.QuestionFormat == QuestionFormatText && question.MsgID != 0 {
		// For text questions, edit the existing message to add edit button
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

/*
sendNewAnswerSummary sends a new message with answer summary and edit button.
Used for radio/checkbox questions or as fallback when editing fails.
*/
func (q *Questionaire) sendNewAnswerSummary(ctx context.Context, b *bot.Bot, question *Question, displayAnswer string, editKB *inline.Keyboard) {
	answerText := fmt.Sprintf("✅ *%s*\n%s",
		helper.EscapeTelegramReserved(question.Text),
		helper.EscapeTelegramReserved(displayAnswer))

	answerMsg, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      q.chatID,
		Text:        answerText,
		ParseMode:   models.ParseModeMarkdown,
		ReplyMarkup: editKB,
	})

	if err != nil {
		// Could optionally log error here if logging system is available
		return
	}

	q.msgIds = append(q.msgIds, answerMsg.ID)
}

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

/*
SetAnswer sets the answer for the question.
*/
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
		return q.validator(answer)
	}
	return nil
}

/*
NewBuilder creates a new Questionaire instance with an optional manager.
*/
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
	}
}

/*
SetManager sets or updates the manager for this questionnaire and returns the updated instance.
*/
func (q *Questionaire) SetManager(m *Manager) *Questionaire {
	q.manager = m
	return q
}

/*
SetContext sets the context for the questionnaire and returns the updated instance.
*/
func (q *Questionaire) SetContext(ctx context.Context) *Questionaire {
	q.ctx = ctx
	return q
}

/*
AddMultipleAnswerQuestion adds a question that expects multiple answers (checkbox style) to the questionnaire.
*/
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

/*
AddQuestion adds a question that expects a single answer (text or radio style) to the questionnaire.
*/
func (q *Questionaire) AddQuestion(key string, text string, choices [][]button.Button, validateFunc func(answer string) error) *Questionaire {
	question := &Question{
		Key:             key,
		Text:            text,
		Choices:         choices,
		ChoicesSelected: make([]string, 0),
		validator:       validateFunc,
		// isTextAnswer:    true,
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

/*
SetOnDoneHandler sets the handler to be called when all questions have been answered.
*/
func (q *Questionaire) SetOnDoneHandler(handler onDoneHandlerFunc) *Questionaire {
	q.onDoneHandler = handler
	return q
}

func (q *Questionaire) SetOnCancelHandler(handler func()) *Questionaire {
	q.onCancelHandler = handler
	return q
}

/*
Done is called when all questions have been answered. It marshals the answers to JSON and calls the onDoneHandler.
*/
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

/*
Show starts the questionnaire, sending the current question to the user and registering with the manager if available.
*/
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
		if curQuestion.QuestionFormat != QuestionFormatCheck &&
			curQuestion.QuestionFormat != QuestionFormatRadio {
			inlineKB.Row()
		}
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

	if isDone := q.Answer("cmd_done", b, q.chatID); isDone {
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
	if isDone := q.Answer(string(data), b, q.chatID); isDone {
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
func (q *Questionaire) Answer(answer string, b *bot.Bot, chatID any) bool {
	curQuestion := q.questions[q.currentQuestionIndex]
	previousQuestionIndex := q.currentQuestionIndex

	if curQuestion.QuestionFormat == QuestionFormatCheck && answer == "cmd_done" {
		// For checkbox questions, "cmd_done" means we're advancing to next question
		q.currentQuestionIndex++
		// Send answer summary for the completed checkbox question
		q.sendAnswerSummary(context.Background(), b, previousQuestionIndex)
	} else if curQuestion.QuestionFormat == QuestionFormatCheck && answer != "cmd_done" {
		if err := curQuestion.Validate(answer); err != nil {

			b.SendMessage(context.Background(), &bot.SendMessageParams{
				ChatID:    chatID,
				Text:      err.Error(),
				ParseMode: models.ParseModeMarkdown,
			})

			q.Show(context.Background(), b, chatID)

			return false
		}

		curQuestion.AddChoiceSelected(answer)
		// For checkbox selections, we don't advance yet, so no answer summary
	} else {
		if err := curQuestion.Validate(answer); err != nil {

			b.SendMessage(context.Background(), &bot.SendMessageParams{
				ChatID:    chatID,
				Text:      err.Error(),
				ParseMode: models.ParseModeMarkdown,
			})

			q.Show(context.Background(), b, chatID)

			return false
		}

		curQuestion.SetAnswer(answer)

		if curQuestion.QuestionFormat != QuestionFormatCheck {
			// For text and radio questions, we advance immediately
			q.currentQuestionIndex++
			// Send answer summary for the completed question
			q.sendAnswerSummary(context.Background(), b, previousQuestionIndex)
		}
	}

	if q.currentQuestionIndex < len(q.questions) {
		q.Show(context.Background(), b, chatID)
	}

	return q.currentQuestionIndex >= len(q.questions)
}

/*
GetResultByte marshals the answers of the questionnaire to JSON bytes.
*/
func GetResultByte(q *Questionaire) ([]byte, error) {

	data, err := json.Marshal(q.GetAnswers())
	return data, err
}
