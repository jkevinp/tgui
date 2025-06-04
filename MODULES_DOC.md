# Module Documentation

---

## button

The `button` module provides utilities for creating buttons and organizing them into grids for Telegram bots.

**Types:**
- `Button`: Represents a button with text, callback data, and an onClick handler.
- `ButtonGrid`: Helps build a grid layout of buttons.

**Public Functions:**
- `New(text string, callbackData string, onClick inline.OnSelect) Button`: Creates a new Button with the specified text, callback data, and onClick handler.
- `NewBuilder() *ButtonGrid`: Creates and returns a new ButtonGrid builder.
- `(bg *ButtonGrid) Row() *ButtonGrid`: Adds a new row to the ButtonGrid and returns the builder for chaining.
- `(bg *ButtonGrid) Add(btn ...Button) *ButtonGrid`: Appends one or more Buttons to the most recent row in the ButtonGrid.
- `(bg *ButtonGrid) Build() [][]Button`: Returns the constructed 2D slice of Buttons from the ButtonGrid.

**Example:**
```go
btn1 := button.New("Yes", "yes", onClickHandler)
btn2 := button.New("No", "no", onClickHandler)
grid := button.NewBuilder().Row().Add(btn1, btn2).Build()
```

---

## datatable

The `datatable` module provides a paginated, filterable data table UI for Telegram bots.

**Types:**
- `DataTable`: Main struct for managing table state, pagination, and filters.
- `DataResult`: Holds the result text, markup, and page count.
- `Filter`, `FilterButton`: For building filter UIs.

**Public Functions:**
- `NewDataResult(text string, replyMarkup [][]button.Button, pagesCount int64) DataResult`: Creates a DataResult with the provided text, reply markup, and pages count.
- `NewErrorDataResult(err error) DataResult`: Creates a DataResult representing an error, with the error message as text.
- `New(b *bot.Bot, itemPerPage int, dataHandlerFunc, manager *questionaire.Manager, filterKeys []string) *DataTable`: Creates and initializes a new DataTable.
- `(d *DataTable) Prefix() string`: Returns the prefix of the DataTable widget.
- `(d *DataTable) Show(ctx context.Context, b *bot.Bot, chatID any, filterInput map[string]interface{}) (*models.Message, error)`: Displays the DataTable using the provided filter input.

**Example:**
```go
dt := datatable.New(bot, 10, dataHandler, manager, []string{"Name", "Age"})
dt.Show(ctx, bot, chatID, nil)
```

---

These modules are designed for easy integration with Telegram bots, supporting interactive UIs with minimal code.

---

## datepicker

The `datepicker` module provides a customizable date picker UI for Telegram bots, allowing users to select dates interactively.

**Types:**
- `DatePicker`: Main struct for managing the date picker state and configuration.
- `DatesMode`: Enum for include/exclude date selection modes.
- `OnSelectHandler`, `OnCancelHandler`, `OnErrorHandler`: Callback types for handling user actions.

**Key Functions:**
- `New`: Creates a new DatePicker with options and handlers.
- `Prefix`: Returns the unique prefix for the widget.

**Example:**
```go
picker := datepicker.New(bot, func(ctx context.Context, bot *bot.Bot, mes models.MaybeInaccessibleMessage, date time.Time) {
    // handle selected date
})
```

---

## dialog

The `dialog` module provides a flexible dialog system for Telegram bots, allowing you to define dialog flows using nodes and buttons.

**Types:**
- `Dialog`: Main struct for managing dialog state and nodes.
- `Node`: Represents a dialog step with text and buttons.
- `OnErrorHandler`: Callback for error handling.

**Key Functions:**
- `New`: Creates a new Dialog with nodes and options.
- `Show`: Displays a dialog node by ID.
- `Prefix`: Returns the unique prefix for the dialog.

**Example:**
```go
dlg := dialog.New(bot, []dialog.Node{
    {ID: "start", Text: "Welcome!", Buttons: []dialog.Button{{Text: "Next", CallbackData: "next"}}},
})
dlg.Show(ctx, bot, chatID, "start")
```

---

## editform

The `editform` module provides a dynamic form editor for Telegram bots, allowing users to edit struct fields interactively.

**Types:**
- `EditForm`: Main struct for managing form state and user input.
- `OnDoneEditHandler`: Callback for handling form submission.

**Key Functions:**
- `New`: Creates a new EditForm for a struct.
- `SetFormatter`: Sets custom formatting and transformation for a field.
- `Show`: Displays the form to the user.

**Example:**
```go
form := editform.New(bot, "Edit User", userStruct, onDoneEdit, nil, chatID, manager)
form.Show(ctx)
```

---

## helper

The `helper` module provides utility functions for string processing.

**Functions:**
- `EscapeTelegramReserved(s string) string`: Escapes Telegram reserved characters in a string for safe display.

**Example:**
```go
escaped := helper.EscapeTelegramReserved("example_text (1.0)")
// Output: example\_text \(1\.0\)
```

---

## keyboard/inline

The `keyboard/inline` module provides an inline keyboard builder for Telegram bots, supporting button actions and callbacks.

**Types:**
- `Keyboard`: Main struct for building and managing inline keyboards.
- `OnSelect`: Callback type for button selection.
- `OnErrorHandler`: Callback for error handling.

**Key Functions:**
- `New`: Creates a new Keyboard with options.
- `Prefix`: Returns the unique prefix for the keyboard.
- `MarshalJSON`: Serializes the keyboard for Telegram API.

**Example:**
```go
kb := inline.New(bot)
kb.Row().Button("Click Me", []byte("data"), onSelectHandler)
```

---

## keyboard/reply

The `keyboard/reply` module provides a reply keyboard builder for Telegram bots, supporting custom keyboard layouts.

**Types:**
- `ReplyKeyboard`: Main struct for building and managing reply keyboards.

**Key Functions:**
- `New`: Creates a new ReplyKeyboard with options.
- `Prefix`: Returns the unique prefix for the keyboard.
- `MarshalJSON`: Serializes the keyboard for Telegram API.

**Example:**
```go
kb := reply.New()
```

---

## menu

The `menu` module provides a simple menu system using reply keyboards for Telegram bots.

**Types:**
- `Menu`: Main struct for managing menu state and keyboard.
- `MenuItem`: Represents a menu item.

**Key Functions:**
- `NewMenuItem`: Creates a new menu item.
- `NewMenu`: Creates a new menu with items.
- `Show`: Displays the menu to the user.

**Example:**
```go
item1 := menu.NewMenuItem("Option 1")
item2 := menu.NewMenuItem("Option 2")
m := menu.NewMenu(bot, "Choose an option:", [][]*menu.MenuItem{{item1, item2}})
m.Show(ctx, bot, chatID)
```

---

## paginator

The `paginator` module provides a paginated list UI for Telegram bots, allowing users to navigate through large lists.

**Types:**
- `Paginator`: Main struct for managing pagination state and data.
- `OnErrorHandler`: Callback for error handling.

**Key Functions:**
- `New`: Creates a new Paginator with data and options.
- `Show`: Displays the current page to the user.
- `Prefix`: Returns the unique prefix for the paginator.

**Example:**
```go
p := paginator.New(bot, []string{"Item 1", "Item 2", "Item 3"})
p.Show(ctx, bot, chatID)
```

---

## parser

The `parser` module provides utilities for parsing struct tags, especially for Telegram UI customization.

**Functions:**
- `ParseTGTags(v interface{}) (map[string]map[string]string, error)`: Parses `tg` and `json` tags from struct fields and returns a map of field names to tag key-value pairs.

**Example:**
```go
type User struct {
    Name string `tg:"label:Name;required" json:"name"`
    Age  int    `tg:"label:Age" json:"age"`
}
tags, err := parser.ParseTGTags(User{})
// tags["name"]["label"] == "Name"
// tags["age"]["label"] == "Age"
```

---

## progress

The `progress` module provides a progress bar UI for Telegram bots, allowing you to display and update progress interactively.

**Types:**
- `Progress`: Main struct for managing progress state and display.
- `OnCancelFunc`: Callback for cancel actions.
- `RenderTextFunc`: Callback for rendering progress text.
- `OnErrorHandler`: Callback for error handling.

**Key Functions:**
- `New`: Creates a new Progress bar with options.
- `Show`: Displays the progress bar to the user.
- `SetValue`: Updates the progress value.
- `Delete`: Removes the progress message.
- `Done`: Cleans up handlers after completion.

**Example:**
```go
progress := progress.New(bot)
progress.Show(ctx, bot, chatID)
progress.SetValue(ctx, bot, 50.0)
```

---

## questionaire

The `questionaire` module provides a flexible questionnaire builder for Telegram bots, supporting multiple question types and validation.

**Types:**
- `Questionaire`: Main struct for managing questions, answers, and flow.
- `Question`: Represents a single question with text, choices, and validation.
- `Manager`: Handles questionnaire sessions for multiple chats.

**Public Functions:**
- `NewBuilder(chatID any, manager *Manager) *Questionaire`: Creates a new Questionaire instance with an optional manager.
- `(q *Questionaire) GetAnswers() map[string]interface{}`: Returns a map of question keys to their answers or selected choices.
- `(q *Questionaire) SetInitialData(data map[string]interface{}) *Questionaire`: Sets the initial data for the questionnaire.
- `(q *Questionaire) SetManager(m *Manager) *Questionaire`: Sets or updates the manager for this questionnaire.
- `(q *Questionaire) SetContext(ctx context.Context) *Questionaire`: Sets the context for the questionnaire.
- `(q *Questionaire) AddMultipleAnswerQuestion(key, text string, choices [][]button.Button, validateFunc func(answer string) error) *Questionaire`: Adds a question that expects multiple answers (checkbox style).
- `(q *Questionaire) AddQuestion(key, text string, choices [][]button.Button, validateFunc func(answer string) error) *Questionaire`: Adds a question that expects a single answer (text or radio style).
- `(q *Questionaire) SetOnDoneHandler(handler onDoneHandlerFunc) *Questionaire`: Sets the handler to be called when all questions have been answered.
- `(q *Questionaire) Done(ctx context.Context, b *bot.Bot, update *models.Update)`: Called when all questions have been answered; marshals answers to JSON and calls the handler.
- `(q *Questionaire) Ask(ctx context.Context, b *bot.Bot, chatID any)`: Starts the questionnaire, sending the current question to the user.
- `(q *Questionaire) Answer(answer string, b *bot.Bot, chatID any) bool`: Processes the user's answer for the current question and advances the questionnaire if appropriate.
- `GetResultByte(q *Questionaire) ([]byte, error)`: Marshals the answers of the questionnaire to JSON bytes.

**Example:**
```go
q := questionaire.NewBuilder(chatID, manager).
    AddQuestion("name", "What is your name?", nil, nil).
    AddQuestion("age", "How old are you?", nil, nil).
    SetOnDoneHandler(func(ctx context.Context, b *bot.Bot, chatID any, answers []byte) error {
        // handle answers
        return nil
    })
q.Ask(ctx, bot, chatID)
```

---

## slider

The `slider` module provides an interactive slider UI for Telegram bots, allowing users to browse through slides with images and text.

**Types:**
- `Slider`: Main struct for managing slides and navigation.
- `Slide`: Represents a single slide with photo and text.
- `OnSelectFunc`, `OnCancelFunc`, `OnErrorFunc`: Callback types for handling actions.

**Key Functions:**
- `New`: Creates a new Slider with slides and options.
- `Show`: Displays the current slide to the user.
- `Prefix`: Returns the unique prefix for the slider.

**Example:**
```go
slides := []slider.Slide{
    {Photo: "https://example.com/image1.png", Text: "Slide 1"},
    {Photo: "https://example.com/image2.png", Text: "Slide 2"},
}
s := slider.New(bot, slides)
s.Show(ctx, bot, chatID)
```

---

## submenu

The `submenu` module provides a submenu system using inline keyboards for Telegram bots.

**Types:**
- `SubMenu`: Main struct for managing submenu state and keyboard.
- `SubMenuItem`: Represents a submenu item.

**Key Functions:**
- `NewSubMenuItem`: Creates a new submenu item.
- `NewSubMenu`: Creates a new submenu with items.
- `Show`: Displays the submenu to the user.

**Example:**
```go
item1 := submenu.NewSubMenuItem("Sub 1", "sub1", onSelect)
item2 := submenu.NewSubMenuItem("Sub 2", "sub2", onSelect)
m := submenu.NewSubMenu(bot, "Choose a submenu:", [][]*submenu.SubMenuItem{{item1, item2}})
m.Show(ctx, bot, chatID)
