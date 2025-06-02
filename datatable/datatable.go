package datatable

import (
	"context"
	"encoding/json"
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

type OnErrorHandler func(err error)

const (
	FILTER    = "🔎 Filter"
	NEXT      = "⏭️ Next"
	BACK      = "⏮️ Back"
	CLOSE     = "❌ Close"
	NODATA    = "No data"
	FILTER_BY = "Filter by"
	CANCEL    = "⬅️ Cancel"
)

type DataTable struct {
	text        string
	replyMarkup [][]button.Button

	prefix              string
	onError             OnErrorHandler
	questionaireManager *questionaire.Manager

	// callbackHandlerID string
	dataHandler dataHandlerFunc

	CtrlBack   button.Button
	CtrlNext   button.Button
	CtrlClose  button.Button
	CtrlFilter button.Button

	filterKeys    []string
	filterMenu    Filter
	currentFilter map[string]interface{}
	filterButtons [][]button.Button

	msgID  any
	chatID any

	b          *bot.Bot
	pagesCount int64
}

func (d *DataTable) calcStartPage() int64 {
	if d.pagesCount < 5 { // 5 is pages buttons count
		return 1
	}
	if int64(d.currentFilter["pageNum"].(float64)) < 3 { // 3 is center page button
		return 1
	}
	if int64(d.currentFilter["pageNum"].(float64)) >= d.pagesCount-2 {
		return d.pagesCount - 4
	}
	return int64(d.currentFilter["pageNum"].(float64)) - 2
}

type DataResult struct {
	Text        string
	ReplyMarkup any
}

// set the struct used to filter the datatable

type dataHandlerFunc func(ctx context.Context, b *bot.Bot, pageSize, pageNum int, filterInput []byte) (text string, replyMarkUp [][]button.Button, maxPage int64)

func New(
	b *bot.Bot,
	itemPerPage int,
	dataHandlerFunc dataHandlerFunc,
	manager *questionaire.Manager,
	filterKeys []string,
) *DataTable {
	prefix := "dt" + bot.RandomString(14)
	fmt.Println("new datatable", prefix)
	p := &DataTable{
		prefix:              prefix,
		onError:             defaultOnError,
		dataHandler:         dataHandlerFunc,
		CtrlBack:            button.Button{Text: BACK, CallbackData: "back"},
		CtrlNext:            button.Button{Text: NEXT, CallbackData: "next"},
		CtrlClose:           button.Button{Text: CLOSE, CallbackData: "close"},
		CtrlFilter:          button.Button{Text: FILTER, CallbackData: "filter"},
		questionaireManager: manager,
		filterKeys:          filterKeys,
		filterMenu:          NewFilter(filterKeys),
		currentFilter:       make(map[string]interface{}),
		b:                   b,
	}

	p.currentFilter["pageSize"] = float64(itemPerPage)
	p.currentFilter["pageNum"] = float64(1)

	filterMenu := button.NewBuilder()

	for _, filterKey := range filterKeys {
		fmt.Println("[datatable] adding filter:", filterKey)
		filterMenu.Row().Add(button.Button{
			Text:         filterKey,
			CallbackData: p.prefix + "filter_" + filterKey,
			OnClick:      p.nagivateCallback,
		})
	}

	if len(filterKeys) > 0 {
		filterMenu.Row().Add(button.New(
			CANCEL,
			p.prefix+"filter_cancel",
			p.nagivateCallback,
		))
	}

	p.filterButtons = filterMenu.Build()

	return p
}

// Prefix returns the prefix of the widget
func (d *DataTable) Prefix() string {
	return d.prefix
}

func defaultOnError(err error) {
	log.Printf("[datatable] [ERROR] %s", err)
}

func (p *DataTable) callbackAnswer(ctx context.Context, b *bot.Bot, callbackQuery *models.CallbackQuery) {
	fmt.Println("callback Answer:")
	ok, err := b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: callbackQuery.ID,
	})
	if err != nil {
		p.onError(err)
		return
	}
	if !ok {
		p.onError(fmt.Errorf("callback answer failed"))
	}
}

func (d *DataTable) nagivateCallback(ctx context.Context, b *bot.Bot, mes models.MaybeInaccessibleMessage, callbackData []byte) {

	fmt.Println("[datatable.nagivateCallback] ", d.prefix, "->", string(callbackData))

	command := strings.TrimPrefix(string(callbackData), d.prefix)

	switch command {
	case "next":
		fmt.Println("[datatable.nagivateCallback] next page")
		d.currentFilter["pageNum"] = d.currentFilter["pageNum"].(float64) + 1
		d.Show(ctx, b, d.chatID, d.getFilterBytes())
	case "back":
		fmt.Println("[datatable.nagivateCallback] back page, current page:", d.currentFilter["pageNum"].(float64))
		// && int64(d.currentFilter["pageNum"].(float64)) <= d.pagesCount
		if d.currentFilter["pageNum"].(float64) > 1 {

			d.currentFilter["pageNum"] = d.currentFilter["pageNum"].(float64) - 1
			fmt.Println("[datatable.nagivateCallback] back page", d.currentFilter["pageNum"].(float64))
			d.Show(ctx, b, d.chatID, d.getFilterBytes())
		}

	case "filter":
		fmt.Println("[datatable.nagivateCallback] filter")

		filterNode := inline.New(d.b, inline.WithPrefix(d.prefix+"filter_"))

		for _, b := range d.filterButtons {
			filterNode.Row()
			for _, btn := range b {

				filterKey := strings.TrimPrefix(btn.CallbackData, d.prefix+"filter_")

				if d.currentFilter[filterKey] != nil {
					btn.Text = fmt.Sprintf("%s: %v", btn.Text, d.currentFilter[filterKey])
				}

				filterNode.Button(
					btn.Text,
					[]byte(btn.CallbackData),
					btn.OnClick,
				)
			}
		}

		_, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: d.chatID,
			// MessageID:   d.msgID.(int),
			Text:        FILTER_BY,
			ParseMode:   models.ParseModeMarkdown,
			ReplyMarkup: filterNode,
		})
		if err != nil {
			d.onError(err)
		}
	case "nop":
		// d.callbackAnswer(ctx, b, update.CallbackQuery)
		fmt.Println("nop")
		return
	case "close":
		fmt.Println("[datatable.nagivateCallback] close")
	case "filter_cancel":
		d.Show(ctx, b, d.chatID, d.getFilterBytes())
	default:

		fmt.Println("[datatable.nagivateCallback] data:", command)
		if strings.HasPrefix(command, "filter_") {

			filterKey := strings.TrimPrefix(command, "filter_")
			//get input from user for the value of filter key

			mapKeysChoice := make(map[string][]string)
			mapKeysChoice[filterKey] = nil

			q := questionaire.NewBuilder(d.chatID, d.questionaireManager).
				AddQuestion(
					filterKey,
					"Enter value for "+filterKey,
					nil,
					nil,
				)

			fun := func(ctx context.Context, b *bot.Bot, chatID any, result []byte) error {

				d.currentFilter["pageNum"] = float64(1)

				var temp map[string]string
				json.Unmarshal(result, &temp)

				fmt.Println("parsing", temp, string(result))

				d.updateFilter(filterKey, temp[filterKey])

				fmt.Println("[datatable]filter conversation:", d.msgID)

				_, err := d.Show(ctx, b, chatID, d.getFilterBytes())

				return err
			}
			q.SetOnDoneHandler(fun)

			q.Ask(ctx, b, d.chatID.(int64))

			return
		} else if strings.HasPrefix(command, "setpage_") {

			fmt.Println("[datatable.nagivateCallback] set page")

			page := strings.TrimPrefix(command, "setpage_")
			pageInt, err := strconv.Atoi(page)
			if err != nil {
				d.onError(err)
				return
			}
			d.currentFilter["pageNum"] = float64(pageInt)

			fmt.Println("[datatable.nagivateCallback] set page", d.currentFilter["pageNum"].(float64))
			d.Show(ctx, b, d.chatID, d.getFilterBytes())

		} else if strings.HasPrefix(command, "remove_filter_") {
			fmt.Println("[datatable.nagivateCallback] remove filter")
			filterKey := strings.TrimPrefix(command, "remove_filter_")
			d.currentFilter[filterKey] = nil
			d.Show(ctx, b, d.chatID, d.getFilterBytes())
		}
	}

}

func (d *DataTable) rebuildControls(chatID any) *bot.SendMessageParams {
	fmt.Println("[datatable] rebuild controls")

	currentPage := int64(d.currentFilter["pageNum"].(float64))

	navigateNode := inline.New(d.b, inline.WithPrefix(d.prefix))

	if d.replyMarkup != nil {
		fmt.Println("[datatable] replyMarkup", d.replyMarkup)
		for _, row := range d.replyMarkup {
			navigateNode.Row()
			for _, btn := range row {
				navigateNode.Button(btn.Text, []byte(d.prefix+btn.CallbackData), func(ctx context.Context, bot *bot.Bot, mes models.MaybeInaccessibleMessage, data []byte) {

					trimmed := strings.TrimPrefix(string(data), d.prefix)

					fmt.Println("[datatable] callback data:", string(data), "trimmed:", trimmed)

					btn.OnClick(ctx, bot, mes, []byte(trimmed))

				})
			}
		}
	}

	navigateNode.Row()
	if currentPage > 1 {
		navigateNode.Button(d.CtrlBack.Text, []byte(d.prefix+d.CtrlBack.CallbackData), d.nagivateCallback)
	}

	if d.pagesCount > 1 {
		startPage := d.calcStartPage()

		for i := startPage; i < startPage+5; i++ {

			text := fmt.Sprintf("%d", i)
			callbackCommand := fmt.Sprintf("%ssetpage_%d", d.prefix, i)

			if i > d.pagesCount {
				break
			}
			if i == currentPage {
				text = "( " + text + " )"
			}

			navigateNode.Button(
				text,
				[]byte(callbackCommand),
				d.nagivateCallback,
			)
		}
	}

	if currentPage < d.pagesCount {
		navigateNode.Button(d.CtrlNext.Text, []byte(d.prefix+d.CtrlNext.CallbackData), d.nagivateCallback)
	}

	if len(d.filterKeys) > 0 {
		navigateNode.Row().Button(d.CtrlFilter.Text, []byte(d.prefix+d.CtrlFilter.CallbackData), d.nagivateCallback)
	}

	if len(d.currentFilter) > 0 {

		for key, value := range d.currentFilter {
			for _, filter := range d.filterKeys {
				if filter == key && d.currentFilter[key] != nil {
					navigateNode.Row().Button(
						fmt.Sprintf("🗑 %s: %v", key, value),
						[]byte(d.prefix+"remove_filter_"+key),
						d.nagivateCallback,
					)
				}
			}

		}
	}

	navigateNode.Row().Button(d.CtrlClose.Text, []byte(d.prefix+d.CtrlClose.CallbackData), d.nagivateCallback)

	params := &bot.SendMessageParams{
		ChatID:      chatID,
		Text:        d.text,
		ParseMode:   models.ParseModeMarkdown,
		ReplyMarkup: navigateNode,
	}

	return params
}

func (d *DataTable) invokeDataHandler(ctx context.Context, b *bot.Bot, pageSize, pageNum int, filterInput []byte) {

	data, replyMarkUp, maxPage := d.dataHandler(ctx, b, pageSize, pageNum, filterInput)
	fmt.Println("[datatable InvokeDataHandler] filter:", string(filterInput), "data:", data, "replyMarkUp:", len(replyMarkUp), "maxPage:", maxPage)
	d.text = data
	d.replyMarkup = replyMarkUp

	if d.replyMarkup == nil && d.text == "" {
		d.text = NODATA
	}

	d.pagesCount = maxPage
}

// show the database using the filterInput(bytes), struct must have pageSize and pageNum
func (d *DataTable) Show(ctx context.Context, b *bot.Bot, chatID any, filterInput []byte) (*models.Message, error) {
	fmt.Println("[datatable] show page , filter:", string(filterInput))
	d.saveFilter(filterInput)
	d.invokeDataHandler(
		ctx,
		b,
		int(d.currentFilter["pageSize"].(float64)),
		int(d.currentFilter["pageNum"].(float64)),
		d.getFilterBytes(),
	)
	params := d.rebuildControls(chatID)
	m, err := b.SendMessage(ctx, params)
	d.msgID = m.ID
	d.chatID = m.Chat.ID
	return m, err
}

func (d *DataTable) saveFilter(filterInput []byte) {

	if d.currentFilter != nil {
		fmt.Println("[datatable] SaveFilter", string(filterInput))
		json.Unmarshal(filterInput, &d.currentFilter)
		fmt.Println("[datatable] Current Filter:", d.currentFilter)
	}

}
func (d *DataTable) updateFilter(key string, value interface{}) {
	d.currentFilter[key] = value
}

func (d *DataTable) getFilterBytes() []byte {
	filterBytes, _ := json.Marshal(d.currentFilter)
	return filterBytes
}

// func (d *DataTable) defaultCallback(ctx context.Context, b *bot.Bot, update *models.Update) {
// 	fmt.Println("[datatable] default callback fired, received:", update.CallbackQuery.Data)
// 	ok, err := b.AnswerCallbackQuery(
// 		ctx,
// 		&bot.AnswerCallbackQueryParams{
// 			CallbackQueryID: update.CallbackQuery.ID,
// 		})

// 	if err != nil {
// 		d.onError(err)
// 	}
// 	if !ok {
// 		d.onError(fmt.Errorf("failed to answer callback query"))
// 	}

// 	data := strings.TrimPrefix(update.CallbackQuery.Data, d.prefix)

// 	fmt.Println("callback data:", data, "prefix:", d.prefix)

// 	if data == "close" {
// 		_, err := b.DeleteMessage(ctx, &bot.DeleteMessageParams{
// 			ChatID:    update.CallbackQuery.Message.Message.Chat.ID,
// 			MessageID: update.CallbackQuery.Message.Message.ID,
// 		})
// 		if err != nil {
// 			d.onError(err)
// 		}
// 		return
// 	} else if data == "filter" {

// 		_, err := b.EditMessageText(ctx, &bot.EditMessageTextParams{
// 			ChatID:      update.CallbackQuery.Message.Message.Chat.ID,
// 			MessageID:   update.CallbackQuery.Message.Message.ID,
// 			Text:        "Filter By:",
// 			ParseMode:   models.ParseModeMarkdown,
// 			ReplyMarkup: d.filterMenu.buildKB(d.prefix),
// 		})
// 		if err != nil {
// 			fmt.Println("filter error")
// 			d.onError(err)
// 		}

// 		return
// 	} else if strings.HasPrefix(data, "filter_") {

// 		filterKey := strings.TrimPrefix(data, "filter_")
// 		//get input from user for the value of filter key

// 		mapKeysChoice := make(map[string][]string)
// 		mapKeysChoice[filterKey] = nil

// 		d.conversationSessions[update.CallbackQuery.Message.Message.Chat.ID] =
// 			questionaire.NewBuilder(d.chatID).
// 				AddQuestion(
// 					filterKey,
// 					"Enter value for "+filterKey,
// 					nil,
// 					nil,
// 				)

// 		fun := func(ctx context.Context, b *bot.Bot, chatID any, result []byte) error {

// 			// d.currentPage = 1 //reset current page to 1
// 			d.currentFilter["pageNum"] = 1

// 			var temp map[string][]string
// 			json.Unmarshal(result, &temp)

// 			fmt.Println("parsing", temp, string(result))

// 			d.UpdateFilter(filterKey, temp[filterKey][0])

// 			fmt.Println("[datatable]filter conversation:", update.CallbackQuery.Message.Message.Chat.ID)
// 			d.InvokeDataHandler(ctx, b, int(d.currentFilter["pageSize"].(float64)), int(d.currentFilter["pageNum"].(float64)), result)
// 			params := d.rebuildControls(d.chatID)
// 			_, errEdit := b.EditMessageText(ctx, &bot.EditMessageTextParams{
// 				ChatID:      d.chatID,
// 				MessageID:   d.msgID.(int),
// 				Text:        d.text,
// 				ParseMode:   models.ParseModeMarkdown,
// 				ReplyMarkup: params.ReplyMarkup,
// 			})
// 			if errEdit != nil {
// 				d.onError(errEdit)
// 				return err
// 			}
// 			return nil
// 		}

// 		d.conversationSessions[update.CallbackQuery.Message.Message.Chat.ID].SetOnDoneHandler(fun)
// 		d.conversationSessions[update.CallbackQuery.Message.Message.Chat.ID].Ask(ctx, b, update.CallbackQuery.Message.Message.Chat.ID)

// 		return
// 	}

// }
