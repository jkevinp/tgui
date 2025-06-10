package datatable

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/jkevinp/tgui/button"
	"github.com/jkevinp/tgui/helper"
	"github.com/jkevinp/tgui/keyboard/inline"
	"github.com/jkevinp/tgui/questionaire"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type OnErrorHandler func(err error)

const (
	FILTER    = "🔎 Filter"
	NEXT      = "➡️"
	BACK      = "⬅️"
	CLOSE     = "❌ Close"
	NODATA    = "No data"
	FILTER_BY = "Filter by"
	CANCEL    = "⬅️ Cancel"

	LASTPAGE  = "%d ⏭️"
	FIRSTPAGE = "%d ⏮️ "

	// Callback data commands and prefixes for internal DataTable controls
	// Standard control button commands
	cbCmdBack   = "back"
	cbCmdNext   = "next"
	cbCmdClose  = "close"
	cbCmdFilter = "filter" // Opens the filter selection menu

	// Parameterized command prefixes
	cbPfxSetPage         = "setpage_"       // Followed by page number, e.g., "setpage_3"
	cbPfxSelectFilterKey = "filter_"        // Followed by filter key, e.g., "filter_category" - to start questionnaire
	cbPfxRemoveFilter    = "remove_filter_" // Followed by filter key, e.g., "remove_filter_category"

	// Specific commands
	cbCmdCancelFilterMenu = "filter_cancel" // Cancels the filter selection menu, returns to table
	cbCmdNop              = "nop"           // No operation, for placeholder buttons if needed
)

type DataTable struct {
	text        string
	replyMarkup [][]button.Button

	prefix              string
	onError             OnErrorHandler
	questionaireManager *questionaire.Manager

	// callbackHandlerID string
	dataHandler     dataHandlerFunc
	onCancelHandler func()

	CtrlBack   button.Button
	CtrlNext   button.Button
	CtrlClose  button.Button
	CtrlFilter button.Button

	filterKeys    []string
	currentFilter map[string]interface{}
	filterButtons [][]button.Button

	msgID  any
	chatID any

	b          *bot.Bot
	pagesCount int64
}

// DataTableBuilder provides a fluent interface for building DataTable instances
type DataTableBuilder struct {
	bot                 *bot.Bot
	itemsPerPage        int
	dataHandler         dataHandlerFunc
	questionaireManager *questionaire.Manager
	filterKeys          []string
	onError             OnErrorHandler
	onCancelHandler     func()
}

// NewBuilder creates a new DataTableBuilder with the required bot instance.
// It sets sensible defaults for optional parameters.
func NewBuilder(b *bot.Bot) *DataTableBuilder {
	if b == nil {
		log.Println("[datatable] [ERROR] NewBuilder: bot instance cannot be nil")
		return nil
	}
	return &DataTableBuilder{
		bot:          b,
		itemsPerPage: 5, // Default items per page
		onError:      defaultOnError,
	}
}

// WithItemsPerPage sets the number of items to display per page.
// If count is not positive, it will be ignored.
func (dtb *DataTableBuilder) WithItemsPerPage(count int) *DataTableBuilder {
	if count > 0 {
		dtb.itemsPerPage = count
	}
	return dtb
}

// WithDataHandler sets the data handler function for fetching data.
// This is a required component for the DataTable to function.
func (dtb *DataTableBuilder) WithDataHandler(handler dataHandlerFunc) *DataTableBuilder {
	dtb.dataHandler = handler
	return dtb
}

// WithFiltering enables filtering capabilities by setting the questionaire manager and filter keys.
// If manager is nil or keys is empty, filtering might be disabled or limited.
func (dtb *DataTableBuilder) WithFiltering(manager *questionaire.Manager, keys []string) *DataTableBuilder {
	dtb.questionaireManager = manager
	dtb.filterKeys = keys
	return dtb
}

// WithOnErrorHandler sets a custom error handler.
// If handler is nil, it will be ignored and the default will be used.
func (dtb *DataTableBuilder) WithOnErrorHandler(handler OnErrorHandler) *DataTableBuilder {
	if handler != nil {
		dtb.onError = handler
	}
	return dtb
}

// WithOnCancelHandler sets a custom cancel handler.
func (dtb *DataTableBuilder) WithOnCancelHandler(handler func()) *DataTableBuilder {
	dtb.onCancelHandler = handler
	return dtb
}

// Build validates the configuration and constructs the DataTable instance.
// It returns an error if any required fields are missing or invalid.
func (dtb *DataTableBuilder) Build() (*DataTable, error) {
	// Validation
	if dtb.bot == nil {
		return nil, errors.New("datatable: Bot instance is required")
	}
	if dtb.dataHandler == nil {
		return nil, errors.New("datatable: DataHandler is required")
	}
	if dtb.itemsPerPage <= 0 {
		return nil, errors.New("datatable: ItemsPerPage must be positive")
	}

	// Construction - using the same logic as the original New() function
	prefix := "dt" + bot.RandomString(14)
	fmt.Println("new datatable", prefix)

	dt := &DataTable{
		b:                   dtb.bot,
		prefix:              prefix,
		onError:             dtb.onError,
		dataHandler:         dtb.dataHandler,
		questionaireManager: dtb.questionaireManager,
		filterKeys:          dtb.filterKeys,
		onCancelHandler:     dtb.onCancelHandler,
		currentFilter:       make(map[string]interface{}),
		// Initialize control buttons
		CtrlBack:   button.Button{Text: BACK, CallbackData: cbCmdBack},
		CtrlNext:   button.Button{Text: NEXT, CallbackData: cbCmdNext},
		CtrlClose:  button.Button{Text: CLOSE, CallbackData: cbCmdClose},
		CtrlFilter: button.Button{Text: FILTER, CallbackData: cbCmdFilter},
	}

	dt.currentFilter["pageSize"] = int64(dtb.itemsPerPage)
	dt.currentFilter["pageNum"] = int64(1)

	// Build filter buttons if filterKeys are provided
	if len(dt.filterKeys) > 0 {
		filterMenu := button.NewBuilder()
		for _, filterKey := range dt.filterKeys {
			fmt.Println("[datatable] adding filter:", filterKey)
			filterMenu.Row().Add(button.Button{
				Text:         filterKey,
				CallbackData: dt.prefix + cbPfxSelectFilterKey + filterKey,
				OnClick:      dt.nagivateCallback,
			})
		}
		filterMenu.Row().Add(button.New(
			CANCEL,
			dt.prefix+cbCmdCancelFilterMenu,
			dt.nagivateCallback,
		))
		dt.filterButtons = filterMenu.Build()
	}

	// Ensure onError is set
	if dt.onError == nil {
		dt.onError = defaultOnError
	}

	return dt, nil
}

func (d *DataTable) calcStartPage() int64 {
	if d.pagesCount < 5 { // 5 is pages buttons count
		return 1
	}
	if (d.currentFilter["pageNum"].(int64)) < 3 { // 3 is center page button
		return 1
	}
	if (d.currentFilter["pageNum"].(int64)) >= d.pagesCount-2 {
		return d.pagesCount - 4
	}
	return (d.currentFilter["pageNum"].(int64)) - 2
}

type DataResult struct {
	Text        string
	ReplyMarkup [][]button.Button
	PagesCount  int64
}

/*
NewDataResult creates a DataResult with the provided text, reply markup, and pages count.
*/
func NewDataResult(text string, replyMarkup [][]button.Button, pagesCount int64) DataResult {
	return DataResult{
		Text:        text,
		ReplyMarkup: replyMarkup,
		PagesCount:  pagesCount,
	}
}

/*
NewErrorDataResult creates a DataResult representing an error, with the error message as text.
*/
func NewErrorDataResult(err error) DataResult {
	return DataResult{
		Text: err.Error(),
		// ReplyMarkup: [][]button.Button{
		// 	{button.Button{Text: CLOSE, CallbackData: "close"}},
		// },
		ReplyMarkup: nil,
		PagesCount:  0,
	}
}

// dataHandlerFunc is a function that handles data retrieval based on the provided context, bot, page size, page number, and filter.
type dataHandlerFunc func(ctx context.Context, b *bot.Bot, pageSize, pageNum int, filter map[string]interface{}) DataResult

func (d *DataTable) SetOnCancelHandler(handler func()) *DataTable {
	d.onCancelHandler = handler
	return d
}

/*
New creates and initializes a new DataTable with the given bot, items per page, data handler, questionaire manager, and filter keys.

Deprecated: Use NewBuilder() instead for a more flexible and readable API.
*/
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
		CtrlBack:            button.Button{Text: BACK, CallbackData: cbCmdBack},
		CtrlNext:            button.Button{Text: NEXT, CallbackData: cbCmdNext},
		CtrlClose:           button.Button{Text: CLOSE, CallbackData: cbCmdClose},
		CtrlFilter:          button.Button{Text: FILTER, CallbackData: cbCmdFilter},
		questionaireManager: manager,
		filterKeys:          filterKeys,
		currentFilter:       make(map[string]interface{}),
		b:                   b,
	}

	p.currentFilter["pageSize"] = int64(itemPerPage)
	p.currentFilter["pageNum"] = int64(1)

	filterMenu := button.NewBuilder()

	for _, filterKey := range filterKeys {
		fmt.Println("[datatable] adding filter:", filterKey)
		filterMenu.Row().Add(button.Button{
			Text:         filterKey,
			CallbackData: p.prefix + cbPfxSelectFilterKey + filterKey,
			OnClick:      p.nagivateCallback,
		})
	}

	if len(filterKeys) > 0 {
		filterMenu.Row().Add(button.New(
			CANCEL,
			p.prefix+cbCmdCancelFilterMenu,
			p.nagivateCallback,
		))
	}

	p.filterButtons = filterMenu.Build()

	return p
}

/*
Prefix returns the prefix of the DataTable widget.
*/
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

	case cbCmdNext:
		d.handleNextPage(ctx, b, mes)
	case cbCmdBack:
		d.handlePreviousPage(ctx, b, mes)
	case cbCmdFilter:
		d.handleShowFilterMenu(ctx, b, mes)
	case cbCmdNop:
		d.handleNop(ctx, b, mes)
	case cbCmdClose:
		d.handleClose(ctx, b, mes)
	case cbCmdCancelFilterMenu:
		d.handleFilterCancel(ctx, b, mes)
	default:
		fmt.Println("[datatable.nagivateCallback] data:", command)
		if strings.HasPrefix(command, cbPfxSelectFilterKey) {
			filterKey := strings.TrimPrefix(command, cbPfxSelectFilterKey)
			d.handleStartFilterQuestionnaire(ctx, b, mes, filterKey)
		} else if strings.HasPrefix(command, cbPfxSetPage) {
			pageStr := strings.TrimPrefix(command, cbPfxSetPage)
			d.handleSetPage(ctx, b, mes, pageStr)
		} else if strings.HasPrefix(command, cbPfxRemoveFilter) {
			filterKey := strings.TrimPrefix(command, cbPfxRemoveFilter)
			d.handleRemoveFilter(ctx, b, mes, filterKey)
		}
	}

}

// handleNextPage processes navigation to the next page
func (d *DataTable) handleNextPage(ctx context.Context, b *bot.Bot, mes models.MaybeInaccessibleMessage) {
	fmt.Println("[datatable.handleNextPage] next page")
	d.currentFilter["pageNum"] = d.currentFilter["pageNum"].(int64) + 1
	d.Show(ctx, b, d.chatID, d.currentFilter)
}

// handlePreviousPage processes navigation to the previous page
func (d *DataTable) handlePreviousPage(ctx context.Context, b *bot.Bot, mes models.MaybeInaccessibleMessage) {
	fmt.Println("[datatable.handlePreviousPage] back page, current page:", d.currentFilter["pageNum"].(int64))
	if d.currentFilter["pageNum"].(int64) > 1 {
		d.currentFilter["pageNum"] = d.currentFilter["pageNum"].(int64) - 1
		fmt.Println("[datatable.handlePreviousPage] back page", d.currentFilter["pageNum"].(int64))
		d.Show(ctx, b, d.chatID, d.currentFilter)
	}
}

// handleShowFilterMenu displays the filter selection menu
func (d *DataTable) handleShowFilterMenu(ctx context.Context, b *bot.Bot, mes models.MaybeInaccessibleMessage) {
	fmt.Println("[datatable.handleShowFilterMenu] filter")

	filterNode := inline.New(d.b, inline.WithPrefix(d.prefix+cbPfxSelectFilterKey))

	for _, filterButtonRow := range d.filterButtons {
		filterNode.Row()
		for _, btn := range filterButtonRow {
			filterKey := strings.TrimPrefix(btn.CallbackData, d.prefix+cbPfxSelectFilterKey)

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
		ChatID:      d.chatID,
		Text:        FILTER_BY,
		ParseMode:   models.ParseModeMarkdown,
		ReplyMarkup: filterNode,
	})
	if err != nil {
		d.onError(err)
	}
}

// handleNop handles no-operation callbacks
func (d *DataTable) handleNop(ctx context.Context, b *bot.Bot, mes models.MaybeInaccessibleMessage) {
	fmt.Println("nop")
	return
}

// handleClose handles the close button callback
func (d *DataTable) handleClose(ctx context.Context, b *bot.Bot, mes models.MaybeInaccessibleMessage) {
	fmt.Println("[datatable.handleClose] close")
	if d.onCancelHandler != nil {
		fmt.Println("[datatable.handleClose] calling onCancelHandler")
		d.onCancelHandler()
	}
}

// handleFilterCancel handles cancelling the filter menu and returning to the table
func (d *DataTable) handleFilterCancel(ctx context.Context, b *bot.Bot, mes models.MaybeInaccessibleMessage) {
	d.Show(ctx, b, d.chatID, d.currentFilter)
}

// handleStartFilterQuestionnaire starts a questionnaire for a specific filter key
func (d *DataTable) handleStartFilterQuestionnaire(ctx context.Context, b *bot.Bot, mes models.MaybeInaccessibleMessage, filterKey string) {
	fmt.Println("[datatable.handleStartFilterQuestionnaire] filter key:", filterKey)

	mapKeysChoice := make(map[string][]string)
	mapKeysChoice[filterKey] = nil

	fun := func(ctx context.Context, b *bot.Bot, chatID any, result map[string]interface{}) error {
		// reset current page on filter change
		d.currentFilter["pageNum"] = int64(1)
		d.updateFilter(filterKey, result[filterKey])
		fmt.Println("[datatable]filter conversation:", d.currentFilter, d.msgID)
		_, err := d.Show(ctx, b, chatID, d.currentFilter)
		return err
	}

	questionaire.NewBuilder(d.chatID, d.questionaireManager).
		AddQuestion(filterKey, "Enter value for "+filterKey, nil, nil).
		SetOnDoneHandler(fun).
		SetAllowEditAnswers(false).
		Show(ctx, b, d.chatID.(int64))
}

// handleSetPage handles navigation to a specific page
func (d *DataTable) handleSetPage(ctx context.Context, b *bot.Bot, mes models.MaybeInaccessibleMessage, pageStr string) {
	fmt.Println("[datatable.handleSetPage] set page")

	pageInt, err := strconv.Atoi(pageStr)
	if err != nil {
		d.onError(err)
		return
	}
	d.currentFilter["pageNum"] = int64(pageInt)

	fmt.Println("[datatable.handleSetPage] set page", d.currentFilter["pageNum"].(int64))
	d.Show(ctx, b, d.chatID, d.currentFilter)
}

// handleRemoveFilter handles removing a specific filter
func (d *DataTable) handleRemoveFilter(ctx context.Context, b *bot.Bot, mes models.MaybeInaccessibleMessage, filterKey string) {
	fmt.Println("[datatable.handleRemoveFilter] remove filter")
	d.currentFilter[filterKey] = nil
	d.Show(ctx, b, d.chatID, d.currentFilter)
}

func (d *DataTable) rebuildControls(chatID any) *bot.SendMessageParams {
	fmt.Println("[datatable] rebuild controls")

	currentPage := int64(d.currentFilter["pageNum"].(int64))
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

	if d.pagesCount > 1 {
		startPage := d.calcStartPage()

		// Show page 1 button if it's not in current navigation range
		if startPage > 1 {
			text := fmt.Sprintf(FIRSTPAGE, 1)
			callbackCommand := fmt.Sprintf("%s%s1", d.prefix, cbPfxSetPage)
			navigateNode.Button(
				text,
				[]byte(callbackCommand),
				d.nagivateCallback,
			)
		}

		if currentPage > 1 {
			navigateNode.Button(d.CtrlBack.Text, []byte(d.prefix+d.CtrlBack.CallbackData), d.nagivateCallback)
		}

		// Show pagination buttons
		for i := startPage; i < startPage+5 && i <= d.pagesCount; i++ {
			text := fmt.Sprintf("%d", i)
			callbackCommand := fmt.Sprintf("%s%s%d", d.prefix, cbPfxSetPage, i)

			if i == currentPage {
				text = "( " + text + " )"
			}

			navigateNode.Button(
				text,
				[]byte(callbackCommand),
				d.nagivateCallback,
			)
		}

		if currentPage < d.pagesCount {
			navigateNode.Button(d.CtrlNext.Text, []byte(d.prefix+d.CtrlNext.CallbackData), d.nagivateCallback)

			// Show last page button if it's not in current navigation range
			lastVisible := startPage + 5 - 1
			if lastVisible < d.pagesCount {
				text := fmt.Sprintf(LASTPAGE, d.pagesCount)
				callbackCommand := fmt.Sprintf("%s%s%d", d.prefix, cbPfxSetPage, d.pagesCount)
				navigateNode.Button(
					text,
					[]byte(callbackCommand),
					d.nagivateCallback,
				)
			}
		}
	}

	// show filter button if there are filter keys
	if len(d.filterKeys) > 0 {
		navigateNode.Row().Button(
			d.CtrlFilter.Text,
			[]byte(d.prefix+d.CtrlFilter.CallbackData),
			d.nagivateCallback,
		)
	}

	// show current filters and allow to remove them
	if len(d.currentFilter) > 0 {

		for key, value := range d.currentFilter {
			for _, filter := range d.filterKeys {
				if filter == key && d.currentFilter[key] != nil {
					navigateNode.Row().Button(
						fmt.Sprintf("🗑 %s: %v", key, value),
						[]byte(d.prefix+cbPfxRemoveFilter+key),
						d.nagivateCallback,
					)
				}
			}

		}
	}

	// show close button
	navigateNode.Row().Button(
		d.CtrlClose.Text,
		[]byte(d.prefix+d.CtrlClose.CallbackData),
		d.nagivateCallback,
	)

	params := &bot.SendMessageParams{
		ChatID:      chatID,
		Text:        helper.EscapeTelegramReserved(d.text),
		ParseMode:   models.ParseModeMarkdown,
		ReplyMarkup: navigateNode,
	}

	return params
}

func (d *DataTable) invokeDataHandler(ctx context.Context, b *bot.Bot, pageSize, pageNum int, filter map[string]interface{}) {

	dataResult := d.dataHandler(ctx, b, pageSize, pageNum, filter)
	fmt.Println("[datatable InvokeDataHandler] filter:", filter, "result:", dataResult)
	d.text = helper.EscapeTelegramReserved(dataResult.Text)
	d.replyMarkup = dataResult.ReplyMarkup

	if d.replyMarkup == nil && d.text == "" {
		d.text = NODATA
	}

	d.pagesCount = dataResult.PagesCount
}

/*
Show displays the DataTable using the provided filter input. The filter input must include pageSize and pageNum.
*/
func (d *DataTable) Show(ctx context.Context, b *bot.Bot, chatID any, filterInput map[string]interface{}) (*models.Message, error) {
	fmt.Println("[datatable] show page , filter:", filterInput)
	d.saveFilter(filterInput)
	d.invokeDataHandler(
		ctx,
		b,
		int(d.currentFilter["pageSize"].(int64)),
		int(d.currentFilter["pageNum"].(int64)),
		d.currentFilter,
	)
	params := d.rebuildControls(chatID)
	m, err := b.SendMessage(ctx, params)
	d.msgID = m.ID
	d.chatID = m.Chat.ID
	return m, err
}

func (d *DataTable) saveFilter(filterInput map[string]interface{}) {

	if d.currentFilter != nil {
		fmt.Println("[datatable] SaveFilter", filterInput)

		if filterInput == nil {
			fmt.Println("[datatable] filterInput is nil, using currentFilter")
			// filterInput = d.currentFilter
		} else {
			fmt.Println("[datatable] filterInput is not nil, updating currentFilter")
			for key, value := range filterInput {
				if value == nil {
					delete(d.currentFilter, key)
				} else {
					if key == "pageSize" || key == "pageNum" {
						// convert value to int64
						if v, ok := value.(int); ok {
							d.updateFilter(key, int64(v))
						} else if v, ok := value.(float64); ok {
							d.updateFilter(key, int64(v))
						} else if v, ok := value.(string); ok {
							if intValue, err := strconv.Atoi(v); err == nil {
								d.updateFilter(key, int64(intValue))
							} else {
								fmt.Println("[datatable] error converting value to int64:", err)
								d.updateFilter(key, v)
							}
						} else {
							fmt.Println("[datatable] unknown type for value, using as is:", value)
							d.updateFilter(key, value)
						}
					} else {
						d.updateFilter(key, value)
					}

				}
			}
			fmt.Println("[datatable] Updated currentFilter:", d.currentFilter)
		}

		// d.currentFilter = filterInput
		// json.Unmarshal(filterInput, &d.currentFilter)
		fmt.Println("[datatable] Current Filter:", d.currentFilter)
	}

}
func (d *DataTable) updateFilter(key string, value interface{}) {
	d.currentFilter[key] = value
}
