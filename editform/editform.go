package editform

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/jkevinp/tgui/button"
	"github.com/jkevinp/tgui/keyboard/inline"
	"github.com/jkevinp/tgui/parser"
	"github.com/jkevinp/tgui/questionaire"
)

const (
	TEXT_FORMAT        = "%s: %v"
	TEXT_FORMAT_EDITED = "🆕 %s: %v"
)

type EditForm struct {
	prefix string
	// targetStruct any

	buttons [][]button.Button

	data map[string]interface{}

	botInstance *bot.Bot

	OnDoneEditHandler OnDoneEditHandler

	text string

	manager *questionaire.Manager

	initialData map[string]interface{}

	targetStructTags map[string]map[string]string

	choices map[string][][]button.Button

	stringFormatter   map[string]func(string) (string, error) //(key,value) return formatted value string
	stringTransformer map[string]func(string) (string, error) //transform input string to string

	chatID any
}

type OnDoneEditHandler func(map[string]interface{}) error

func (f *EditForm) SetFormatter(key string, formatFunc func(string) (string, error), transformFunc func(string) (string, error)) *EditForm {
	f.stringFormatter[key] = formatFunc
	f.stringTransformer[key] = transformFunc
	return f
}

func New(
	b *bot.Bot, //bot instance
	text string, // edit form text
	targetStruct any, //struct to edit
	onDoneEdit OnDoneEditHandler, // onDoneEdit function
	choices map[string][][]button.Button,
	chatID any,
	manager *questionaire.Manager, //conversation sessions

) *EditForm {
	prefix := "ef" + bot.RandomString(14)
	fmt.Println("new editform", prefix)

	f := EditForm{
		prefix:            prefix,
		data:              make(map[string]interface{}),
		botInstance:       b,
		OnDoneEditHandler: onDoneEdit,
		text:              text,
		manager:           manager,
		initialData:       make(map[string]interface{}),
		choices:           choices,
		chatID:            chatID,
		stringFormatter:   make(map[string]func(string) (string, error)),
		stringTransformer: make(map[string]func(string) (string, error)),
	}

	parsedTags, err := parser.ParseTGTags(targetStruct)

	fmt.Println("parsedTags:", parsedTags, err)

	f.targetStructTags = parsedTags

	dataBytes, err := json.Marshal(targetStruct)

	if err != nil {
		fmt.Println(err)
	}

	if err := json.Unmarshal(dataBytes, &f.data); err != nil {
		fmt.Println(err)
	}

	for key, val := range f.data {
		fmt.Println("[editform] key:", key, "val:", val, "type:", reflect.TypeOf(val))
		//add cancel for choices
		btnCancel := []button.Button{
			{
				Text:         "❌ Cancel",
				CallbackData: prefix + "cancel_" + key,
				OnClick: func(ctx context.Context, b *bot.Bot, mes models.MaybeInaccessibleMessage, callbackData []byte) {
					fmt.Println("[EditForm] cancel choice for key:", key)
					b.SendMessage(ctx, &bot.SendMessageParams{
						ChatID: mes.Message.Chat.ID,
						Text:   "Cancelled choice for " + key,
					})
				},
			},
		}

		if f.choices[key] != nil {
			f.choices[key] = append(f.choices[key], btnCancel)
		} else {
			f.choices[key] = [][]button.Button{btnCancel}
		}

	}

	for key, val := range f.data {
		f.initialData[key] = val
	}

	return &f
}

func (f *EditForm) rebuildControls() {
	editForm := button.NewBuilder()

	for key, _ := range f.data {

		tgTags := f.targetStructTags[key]
		if tgTags["noedit"] == "true" {
			fmt.Println("[editform] skipping key:", key)
			continue
		}

		fmtToUse := TEXT_FORMAT

		if f.initialData[key] != f.data[key] {
			fmtToUse = TEXT_FORMAT_EDITED
		}

		fmt.Println("[editform] adding key:", key)

		value := f.data[key]
		var err error

		if f.stringFormatter[key] != nil {
			value, err = f.stringFormatter[key](fmt.Sprintf("%v", value))

			if err != nil {
				f.botInstance.SendMessage(context.Background(), &bot.SendMessageParams{
					ChatID: f.chatID,
					Text:   err.Error(),
				})
				return
			}

			fmt.Println("[editform] adding formatter for key:", key, "new value:", value)
		}

		editForm.Row().Add(button.Button{
			Text:         fmt.Sprintf(fmtToUse, key, value),
			CallbackData: f.prefix + "edit_" + key,
			OnClick:      f.editCallback,
		})
	}

	editForm.Row().Add(button.Button{
		Text:         "✅ Done",
		CallbackData: f.prefix + "done",
		OnClick:      f.editCallback,
	}).Add(button.Button{
		Text:         "❌ Cancel",
		CallbackData: f.prefix + "cancel",
		OnClick:      f.editCallback,
	})

	f.buttons = editForm.Build()
}

func (f *EditForm) editCallback(ctx context.Context, b *bot.Bot, mes models.MaybeInaccessibleMessage, callbackData []byte) {
	fmt.Println("[EditForm.editCallback] ", f.prefix, "->", string(callbackData))

	command := strings.TrimPrefix(string(callbackData), f.prefix)

	switch command {
	case "done":
		fmt.Println("[EditForm.editCallback] done", f.data)
		if err := f.OnDoneEditHandler(f.data); err != nil {
			fmt.Println(err)
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: mes.Message.Chat.ID,
				Text:   err.Error(),
			})
		}
	default:
		if strings.HasPrefix(command, "edit_") {
			fmt.Println("[EditForm.editCallback] edit", command)
			key := strings.TrimPrefix(command, "edit_")

			q := questionaire.NewBuilder(mes.Message.Chat.ID, f.manager).
				SetOnDoneHandler(func(ctx context.Context, b *bot.Bot, chatID any, req map[string]interface{}) error {

					// var req map[string]interface{}
					// if err := json.Unmarshal(answersByte, &req); err != nil {
					// 	return err
					// }

					fmt.Println("[EditForm.editCallback] received answers:", req)

					answer := req[key]
					if answer == nil {
						return fmt.Errorf("no answer for key: %s", key)
					}

					if answer != f.prefix+"cancel_"+key {
						if f.stringTransformer[key] != nil {
							var err error
							req[key], err = f.stringTransformer[key](req[key].(string))
							if err != nil {
								return err
							}
						}

						f.data[key] = req[key]

					}

					f.Show(ctx)

					return nil
				})

			if f.choices[key] != nil {
				q.AddQuestion(
					key,
					"Select value for "+key+" or enter new value: ",
					f.choices[key],
					nil,
				)
			} else {
				q.AddQuestion(key, "Enter new value for: "+key, nil, nil)
			}

			q.Ask(ctx, b, mes.Message.Chat.ID)

		}
	}
}

func (f *EditForm) Show(ctx context.Context) (*models.Message, error) {

	f.rebuildControls()

	filterNode := inline.New(f.botInstance, inline.WithPrefix(f.prefix))

	for _, b := range f.buttons {
		filterNode.Row()
		for _, btn := range b {
			filterNode.Button(btn.Text, []byte(btn.CallbackData), btn.OnClick)
		}
	}
	return f.botInstance.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      f.chatID,
		Text:        f.text,
		ReplyMarkup: filterNode,
	})
}
