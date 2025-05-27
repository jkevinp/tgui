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

type onDoneHandlerFunc func(ctx context.Context, b *bot.Bot, chatID any, answersByte []byte) error

type Questionaire struct {
	questions            []*Question
	currentQuestionIndex int
	onDoneHandler        onDoneHandlerFunc

	callbackID string
	msgIds     []int

	chatID any

	ctx context.Context

	InitialData map[string]interface{}
}

func (q *Questionaire) GetAnswers() map[string]interface{} {
	answers := make(map[string]interface{})
	for _, question := range q.questions {
		if question.isTextAnswer {
			answers[question.Key] = question.Answer
		} else {
			answers[question.Key] = question.ChoicesSelected
		}
	}

	for key, value := range q.InitialData {
		answers[key] = value
	}

	return answers
}

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
	isTextAnswer    bool
}

func (q *Question) SetAnswer(answer string) {
	q.Answer = answer
}

func (q *Question) AddChoiceSelected(answer string) {
	q.ChoicesSelected = append(q.ChoicesSelected, answer)
}

func (q *Question) Validate(answer string) error {
	if q.validator != nil {
		return q.validator(answer)
	}
	return nil
}

func NewBuilder(chatID any) *Questionaire {
	fmt.Println("new question builder:", chatID)
	return &Questionaire{
		questions:            make([]*Question, 0),
		callbackID:           "qs" + bot.RandomString(14),
		chatID:               chatID,
		currentQuestionIndex: 0,
		onDoneHandler:        nil,
		msgIds:               make([]int, 0),
	}
}

func (q *Questionaire) SetContext(ctx context.Context) *Questionaire {
	q.ctx = ctx
	return q
}

// creates a question that expects array of answer
func (q *Questionaire) AddMultipleAnswerQuestion(key string, text string, choices [][]button.Button, validateFunc func(answer string) error) *Questionaire {
	question := &Question{
		Key:             key,
		Text:            text,
		Choices:         make([][]button.Button, 0),
		ChoicesSelected: make([]string, 0),
		validator:       validateFunc,
		isTextAnswer:    false,
	}

	question.Choices = choices

	q.questions = append(q.questions, question)

	fmt.Println("added question:", question)

	return q
}

// creatte a question that expects a text answer
func (q *Questionaire) AddQuestion(key string, text string, choices [][]button.Button, validateFunc func(answer string) error) *Questionaire {
	question := &Question{
		Key:             key,
		Text:            text,
		Choices:         choices,
		ChoicesSelected: make([]string, 0),
		validator:       validateFunc,
		isTextAnswer:    true,
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

// will be called once all questions have been asked
func (q *Questionaire) SetOnDoneHandler(handler onDoneHandlerFunc) *Questionaire {
	q.onDoneHandler = handler
	return q
}

// Called when all questions have been asked, will pass answer struct marshalled to json bytes
func (q *Questionaire) Done(ctx context.Context, b *bot.Bot, update *models.Update) {

	if q.ctx != nil {
		ctx = mergectx.Join(ctx, q.ctx)
	}

	result, err := GetResultByte(q)

	if err != nil {
		fmt.Println("error getting result:", err)
		return
	}
	fmt.Println("result of questionaire:", string(result))

	if err := q.onDoneHandler(ctx, b, q.chatID, result); err != nil {
		fmt.Println("error calling onDoneHandler:", err)
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

func (q *Questionaire) Ask(ctx context.Context, b *bot.Bot, chatID any) {

	curQuestion := q.questions[q.currentQuestionIndex]
	fmt.Println("[question] -> ", q.callbackID, "asking question about:", curQuestion, "choices:", q.questions[q.currentQuestionIndex].Choices)

	params := &bot.SendMessageParams{
		ChatID:    chatID,
		Text:      "✒️" + helper.EscapeTelegramReserved(curQuestion.Text),
		ParseMode: models.ParseModeMarkdown,
	}

	if len(curQuestion.Choices) > 0 && !curQuestion.isTextAnswer {
		inlineKB := inline.New(b, inline.WithPrefix(q.callbackID))

		for _, choiceRow := range q.questions[q.currentQuestionIndex].Choices {
			inlineKB.Row()

			for _, i := range choiceRow {
				inlineKB.Button(i.Text, []byte(i.CallbackData), q.onInlineKeyboardSelect)
			}

		}
		if !curQuestion.isTextAnswer {
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
	m, _ := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: q.chatID,
		Text:   "You selected: " + string(data),
	})

	q.msgIds = append(q.msgIds, m.ID)

	// curQuestion := q.questions[q.currentQuestionIndex]

	if isDone := q.Answer(string(data), b, q.chatID); isDone {
		fmt.Println("isDone:", isDone)
		q.Done(ctx, b, nil)
	}

}

func (q *Questionaire) Answer(answer string, b *bot.Bot, chatID any) bool {
	fmt.Println("answer:", answer)
	curQuestion := q.questions[q.currentQuestionIndex]

	if !curQuestion.isTextAnswer && answer == "cmd_done" {
		q.currentQuestionIndex++
	} else if !curQuestion.isTextAnswer && answer != "cmd_done" {
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

		if curQuestion.isTextAnswer {
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

func GetResultByte(q *Questionaire) ([]byte, error) {

	data, err := json.Marshal(q.GetAnswers())
	return data, err
}
