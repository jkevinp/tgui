package questionaire

import (
	"context"
	"encoding/json"
	"fmt"

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

type Questionaire struct {
	questions            []*Question
	currentQuestionIndex int
	onDoneHandler        onDoneHandlerFunc

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

// func NewByKeys(b *bot.Bot, mapKeysChoice map[string][]string) *Questionaire {

// 	keys := make([]string, len(mapKeysChoice))
// 	i := 0

// 	choices := make([][]string, len(mapKeysChoice))
// 	for key := range mapKeysChoice {
// 		keys[i] = key
// 		if mapKeysChoice[key] != nil {
// 			choices[i] = mapKeysChoice[key]
// 		} else {
// 			choices[i] = []string{}
// 		}
// 		i++
// 	}

// 	return &Questionaire{

// 		currentQuestionIndex: 0,
// 		resultFieldNames:     keys,
// 		choices:              choices,
// 		answersTemp:          make(map[string]string),
// 		callbackID:           "qs" + bot.RandomString(14),
// 	}
// }

// Asks question about each field in the struct
// func NewByStruct(b *bot.Bot, answerStruct any) *Questionaire {
// 	resTags, _ := parser.ParseTGTags(answerStruct)

// 	fieldNames := make([]string, len(resTags))
// 	i := 0
// 	for key := range resTags {
// 		// fmt.Println("key:", key)
// 		fieldNames[i] = key
// 		i++
// 	}

// 	fmt.Println("creating questionaire w/ field names:", fieldNames)

// 	return &Questionaire{
// 		currentQuestionIndex: 0,
// 		resultFieldNames:     fieldNames,
// 		answersTemp:          make(map[string]string),
// 		resultStruct:         answerStruct,
// 		callbackID:           "qs" + bot.RandomString(14),
// 	}
// }

/*
SetOnDoneHandler sets the handler to be called when all questions have been answered.
*/
func (q *Questionaire) SetOnDoneHandler(handler onDoneHandlerFunc) *Questionaire {
	q.onDoneHandler = handler
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

	q = nil

}

/*
Ask starts the questionnaire, sending the current question to the user and registering with the manager if available.
*/
func (q *Questionaire) Ask(ctx context.Context, b *bot.Bot, chatID any) {
	curQuestion := q.questions[q.currentQuestionIndex]
	fmt.Println("[question] -> ", q.callbackID, "asking question about:", curQuestion, "choices:", q.questions[q.currentQuestionIndex].Choices)

	if q.manager != nil && q.chatID != nil {
		q.manager.Add(q.chatID.(int64), q)
	}

	params := &bot.SendMessageParams{
		ChatID:    chatID,
		Text:      "✒️" + helper.EscapeTelegramReserved(curQuestion.Text),
		ParseMode: models.ParseModeMarkdown,
	}

	if len(curQuestion.Choices) > 0 && curQuestion.QuestionFormat != QuestionFormatText {
		inlineKB := inline.New(b, inline.WithPrefix(q.callbackID))

		// Add selected choices
		for _, choiceRow := range curQuestion.GetSelectedChoices() {
			inlineKB.Row()
			for _, i := range choiceRow {
				inlineKB.Button(
					"✅ "+helper.EscapeTelegramReserved(i.Text),
					[]byte(i.CallbackData),
					q.onInlineKeyboardUnSelect,
				)
			}
		}

		// Add unselected choices
		for _, choiceRow := range q.questions[q.currentQuestionIndex].GetUnselectedChoices() {
			inlineKB.Row()

			for _, i := range choiceRow {
				inlineKB.Button(
					"☑️ "+helper.EscapeTelegramReserved(i.Text),
					[]byte(i.CallbackData),
					q.onInlineKeyboardSelect,
				)
			}

		}
		if curQuestion.QuestionFormat == QuestionFormatCheck {
			inlineKB.Row().Button("✅", []byte("cmd_done"), q.onDoneChoosing)
		}

		params.ReplyMarkup = inlineKB
	}

	fmt.Println("reply markup:", params.ReplyMarkup)

	m, err := b.SendMessage(ctx, params)

	if err == nil {
		q.msgIds = append(q.msgIds, m.ID)
	} else {
		fmt.Println("error sending message:", err)
	}

}

func (q *Questionaire) onDoneChoosing(ctx context.Context, b *bot.Bot, mes models.MaybeInaccessibleMessage, data []byte) {
	b.DeleteMessage(ctx, &bot.DeleteMessageParams{
		ChatID:    q.chatID,
		MessageID: mes.Message.ID,
	})

	fmt.Println("cmd_done")

	if isDone := q.Answer("cmd_done", b, q.chatID); isDone {
		fmt.Println("isDone:", isDone)
		q.Done(ctx, b, nil)
	}

}

func (q *Questionaire) onInlineKeyboardSelect(ctx context.Context, b *bot.Bot, mes models.MaybeInaccessibleMessage, data []byte) {
	// m, _ := b.SendMessage(ctx, &bot.SendMessageParams{
	// 	ChatID: q.chatID,
	// 	Text:   "You selected: " + string(data),
	// })

	// q.msgIds = append(q.msgIds, m.ID)

	// curQuestion := q.questions[q.currentQuestionIndex]

	if isDone := q.Answer(string(data), b, q.chatID); isDone {
		fmt.Println("isDone:", isDone)
		q.Done(ctx, b, nil)
	}

}

func (q *Questionaire) onInlineKeyboardUnSelect(ctx context.Context, b *bot.Bot, mes models.MaybeInaccessibleMessage, data []byte) {

	curQuestion := q.questions[q.currentQuestionIndex]
	fmt.Println("unselecting choice:", string(data), "for question:", curQuestion.Key)

	if curQuestion.QuestionFormat == QuestionFormatCheck {
		for i, selectedChoice := range curQuestion.ChoicesSelected {
			if selectedChoice == string(data) {
				curQuestion.ChoicesSelected = append(curQuestion.ChoicesSelected[:i], curQuestion.ChoicesSelected[i+1:]...)
				break
			}
		}
	}

	q.Ask(ctx, b, q.chatID)

}

/*
Answer processes the user's answer for the current question and advances the questionnaire if appropriate.
Returns true if all questions have been answered.
*/
func (q *Questionaire) Answer(answer string, b *bot.Bot, chatID any) bool {
	fmt.Println("answer:", answer)
	curQuestion := q.questions[q.currentQuestionIndex]

	if curQuestion.QuestionFormat == QuestionFormatCheck && answer == "cmd_done" {
		q.currentQuestionIndex++
	} else if curQuestion.QuestionFormat == QuestionFormatCheck && answer != "cmd_done" {
		if err := curQuestion.Validate(answer); err != nil {

			b.SendMessage(context.Background(), &bot.SendMessageParams{
				ChatID:    chatID,
				Text:      err.Error(),
				ParseMode: models.ParseModeMarkdown,
			})

			q.Ask(context.Background(), b, chatID)

			return false
		}

		curQuestion.AddChoiceSelected(answer)
	} else {
		if err := curQuestion.Validate(answer); err != nil {

			b.SendMessage(context.Background(), &bot.SendMessageParams{
				ChatID:    chatID,
				Text:      err.Error(),
				ParseMode: models.ParseModeMarkdown,
			})

			q.Ask(context.Background(), b, chatID)

			return false
		}

		curQuestion.SetAnswer(answer)

		if curQuestion.QuestionFormat != QuestionFormatCheck {
			q.currentQuestionIndex++
		}
	}

	// if err := q.TestInput(); err != nil {
	// 	q.answersTemp[curQuestion] = ""

	// 	q.Ask(context.Background(), b, chatID)

	// 	return false
	// }

	if q.currentQuestionIndex < len(q.questions) {
		q.Ask(context.Background(), b, chatID)
	}

	return q.currentQuestionIndex >= len(q.questions)
}

// func (q *Questionaire) TestInput() error {
// 	data, err := json.Marshal(q.answersTemp)
// 	if err != nil {
// 		return err
// 	}

// 	if err := json.Unmarshal(data, &q.resultStruct); err != nil {
// 		return err
// 	}

// 	return nil
// }

/*
GetResultByte marshals the answers of the questionnaire to JSON bytes.
*/
func GetResultByte(q *Questionaire) ([]byte, error) {

	data, err := json.Marshal(q.GetAnswers())
	return data, err
}
