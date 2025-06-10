# Questionnaire Component

The `questionaire` component for `tgui` allows you to guide a Telegram bot user through a sequence of questions, collect their answers, and process them. It supports text input, single-choice (radio button style), and multiple-choice (checkbox style) questions.

## Quick Start

**Try the live demo:** You can test the questionnaire component right now by messaging [@TGUIEXbot](https://t.me/TGUIEXbot) and sending `/survey`.

To run the example locally:
1. Navigate to `/home/joketa/tgui/examples/`
2. The `.env` file is already configured with a bot token
3. Run: `go run questionaire_example.go`
4. Send `/survey` to the bot to test the questionnaire functionality

### Enhanced Example Features

The included example (`questionaire_example.go`) showcases various ButtonGrid patterns:

- **Text Input**: Name and experience questions with validation
- **Radio Buttons**: Age group selection with single column layout
- **2×2 Grid**: Interest selection with organized checkbox layout
- **Quick Choices**: Yes/No developer question using `QuickChoices`
- **Paired Layout**: Programming language selection using `QuickPairedChoices`
- **Custom Rating**: Star rating system with custom emoji layout
- **Smart Results**: Beautifully formatted result summary with all answers

The example demonstrates 7 different question types and ButtonGrid patterns, making it a comprehensive reference for real-world usage.

## Edit Functionality

The questionnaire includes a powerful edit feature that allows users to go back and modify their previous answers:

- **Automatic Edit Buttons**: After answering each question, previous questions are automatically updated with an edit button showing the user's answer
- **User-Friendly Display**: Edit buttons show human-readable text rather than internal data:
  - Text answers: Shows the actual text entered
  - Radio selections: Shows the selected option text (e.g., "Edit: Under 18")
  - Checkbox selections: Shows the first selection plus count (e.g., "Edit: Technology + 2 more")
- **Smart Reset**: When going back to edit, all subsequent questions are automatically cleared and reset
- **Seamless Flow**: Users can navigate back and forth through questions without losing progress on earlier steps

## Core Concepts

*   **`Questionaire`**: The main object that manages a series of questions for a specific chat. It handles the flow, displays questions, and collects answers.
*   **`Question`**: Represents an individual question within a `Questionaire`. It holds the question text, possible choices (if any), the format of the question, and an optional validation function.
*   **`Manager`**: A session manager for `Questionaire` instances. It keeps track of active questionnaires for different chat IDs and routes incoming text messages from users to the appropriate active `Questionaire`.
*   **`QuestionFormat`**: Defines the type of question:
    *   `QuestionFormatText`: User provides a free-text answer.
    *   `QuestionFormatRadio`: User selects one option from a list of buttons.
    *   `QuestionFormatCheck`: User selects one or more options from a list of buttons.

## Using ButtonGrid for Choices

The questionnaire supports the powerful `ButtonGrid` builder pattern for creating clean, organized choice layouts. This is the **recommended way** to create choices as it provides better maintainability and flexibility than manual button creation.

### ButtonGrid Builder Methods

#### Basic Methods:
- **`NewBuilder()`**: Creates a new ButtonGrid builder
- **`Row()`**: Starts a new row in the grid
- **`Add(button.Button)`**: Adds a button to the current row
- **`Build()`**: Returns the final `[][]button.Button` for use in questions

#### Convenience Methods:
- **`Choice(text)`**: Adds a button with auto-generated callback data
- **`ChoiceWithData(text, callbackData)`**: Adds a button with custom callback data
- **`SingleChoice(text)`**: Creates a new row with one choice (perfect for radio buttons)
- **`SingleChoiceWithData(text, data)`**: Creates a new row with custom callback data

### Basic ButtonGrid Usage

```go
import "github.com/jkevinp/tgui/button"

// Simple single choices (radio buttons) - each on its own row
ageChoices := button.NewBuilder().
    SingleChoiceWithData("Under 18", "age_under_18").
    SingleChoiceWithData("18-30", "age_18_30").
    SingleChoiceWithData("31-45", "age_31_45").
    SingleChoiceWithData("Over 45", "age_over_45").
    Build()

// 2x2 grid layout (great for checkboxes)
topicChoices := button.NewBuilder().
    Row().
    ChoiceWithData("Technology", "topic_tech").
    ChoiceWithData("Sports", "topic_sports").
    Row().
    ChoiceWithData("Music", "topic_music").
    ChoiceWithData("Travel", "topic_travel").
    Build()

// 3-column rating scale
ratingChoices := button.NewBuilder().
    Row().
    ChoiceWithData("⭐", "1").
    ChoiceWithData("⭐⭐", "2").
    ChoiceWithData("⭐⭐⭐", "3").
    Row().
    ChoiceWithData("⭐⭐⭐⭐", "4").
    ChoiceWithData("⭐⭐⭐⭐⭐", "5").
    Build()

// Auto-generated callback data (text becomes callback)
simpleChoices := button.NewBuilder().
    Row().
    Choice("Option A").    // callback: "option_a"
    Choice("Option B").    // callback: "option_b"
    Build()
```

### Quick Helper Functions

For common patterns, use these convenient helper functions:

```go
// Simple Yes/No choices (single column)
yesNoChoices := button.QuickChoices("Yes", "No")

// Paired layout (2 per row) - perfect for even number of options
languageChoices := button.QuickPairedChoices(
    "Go", "Python", "JavaScript", "Java", "C++", "Rust")

// Custom callback data with explicit pairs
customChoices := button.QuickChoicesWithData(
    "Display Text", "callback_data",
    "Another Option", "another_callback",
    "Third Option", "third_option")
```

### Advanced Layout Examples

```go
// Mixed row sizes for complex layouts
complexChoices := button.NewBuilder().
    Row().
    ChoiceWithData("Priority 1", "p1").
    Row().
    ChoiceWithData("Normal", "normal").
    ChoiceWithData("Low", "low").
    Row().
    ChoiceWithData("Not Important", "ni").
    Build()

// Emoji-based choices with custom layout
feedbackChoices := button.NewBuilder().
    Row().
    ChoiceWithData("😍 Love it", "love").
    ChoiceWithData("😊 Like it", "like").
    Row().
    ChoiceWithData("😐 Neutral", "neutral").
    Row().
    ChoiceWithData("😞 Dislike", "dislike").
    ChoiceWithData("😡 Hate it", "hate").
    Build()
```

### Complete Example Usage

```go
// Create questionnaire with ButtonGrid choices
q := questionaire.NewBuilder(chatID, manager).
    SetOnDoneHandler(handleResults).
    SetOnCancelHandler(handleCancel)

// Text question (no choices needed)
q.AddQuestion("name", "What's your name?", nil, validateNonEmpty)

// Radio question with ButtonGrid
ageChoices := button.NewBuilder().
    SingleChoiceWithData("Under 18", "age_under_18").
    SingleChoiceWithData("18-30", "age_18_30").
    SingleChoiceWithData("31-45", "age_31_45").
    SingleChoiceWithData("Over 45", "age_over_45").
    Build()
q.AddQuestion("age_group", "Which age group do you belong to?", ageChoices, nil)

// Checkbox question with 2x2 grid
interestChoices := button.NewBuilder().
    Row().
    ChoiceWithData("Technology", "tech").
    ChoiceWithData("Sports", "sports").
    Row().
    ChoiceWithData("Music", "music").
    ChoiceWithData("Travel", "travel").
    Build()
q.AddMultipleAnswerQuestion("interests",
    "Which topics interest you? (Select multiple, then click Done)",
    interestChoices, nil)

// Simple Yes/No with QuickChoices
developerChoices := button.QuickChoices("Yes", "No")
q.AddQuestion("is_developer", "Are you a software developer?", developerChoices, nil)

// Start the questionnaire
q.Show(ctx, b, chatID)
```

### ButtonGrid Benefits & Best Practices

#### Why Use ButtonGrid?

1. **Cleaner Code**: Builder pattern is more readable than manual array creation
2. **Flexible Layouts**: Easily create 1×n, 2×2, 3×2, or custom arrangements
3. **Less Boilerplate**: Quick helpers eliminate repetitive code
4. **Maintainable**: Easy to modify layouts without restructuring arrays
5. **Type Safety**: Builder pattern prevents malformed button arrays
6. **Consistent**: Standardized approach across your application

#### Best Practices:

**For Radio Questions (Single Selection):**
```go
// ✅ Good: Use SingleChoiceWithData for clear single-column layout
choices := button.NewBuilder().
    SingleChoiceWithData("Yes", "yes").
    SingleChoiceWithData("No", "no").
    SingleChoiceWithData("Maybe", "maybe").
    Build()

// ❌ Avoid: Manual creation for simple cases
choices := [][]button.Button{
    {button.Button{Text: "Yes", CallbackData: "yes"}},
    {button.Button{Text: "No", CallbackData: "no"}},
}
```

**For Checkbox Questions (Multiple Selection):**
```go
// ✅ Good: Use grid layout for better visual organization
choices := button.NewBuilder().
    Row().
    ChoiceWithData("Tech", "tech").
    ChoiceWithData("Sports", "sports").
    Row().
    ChoiceWithData("Music", "music").
    ChoiceWithData("Travel", "travel").
    Build()

// ✅ Also good: Use QuickPairedChoices for even numbers
choices := button.QuickPairedChoices("Tech", "Sports", "Music", "Travel")
```

**For Rating/Scale Questions:**
```go
// ✅ Good: Custom layout for visual appeal
ratingChoices := button.NewBuilder().
    Row().
    ChoiceWithData("1⭐", "1").
    ChoiceWithData("2⭐", "2").
    ChoiceWithData("3⭐", "3").
    Row().
    ChoiceWithData("4⭐", "4").
    ChoiceWithData("5⭐", "5").
    Build()
```

**For Simple Yes/No Questions:**
```go
// ✅ Best: Use QuickChoices helper
choices := button.QuickChoices("Yes", "No")

// ✅ Also good: Explicit if you need custom callback data
choices := button.NewBuilder().
    SingleChoiceWithData("Accept", "accept").
    SingleChoiceWithData("Decline", "decline").
    Build()
```

#### Layout Guidelines:

- **1 column**: Perfect for radio buttons with many options
- **2 columns**: Great for checkboxes, pairs of options
- **3+ columns**: Use sparingly, only for compact items (emojis, numbers)
- **Mixed layouts**: Combine different row sizes for emphasis

## Configuration Options

### Edit Functionality Control

You can control whether users can edit their previous answers using `SetAllowEditAnswers()`:

```go
// Default behavior - users can edit previous answers
q := questionaire.NewBuilder(chatID, manager).
    SetAllowEditAnswers(true)

// Disable editing - no edit buttons will be shown
q := questionaire.NewBuilder(chatID, manager).
    SetAllowEditAnswers(false)
```

When `SetAllowEditAnswers(false)` is used:
- ✅ Answered questions will show as completed (with checkmark)
- ❌ No "◀️ Edit" buttons will be displayed
- ❌ Users cannot go back to modify previous answers
- ✅ Questionnaire flow becomes linear and faster

This is useful for:
- **Surveys where answers shouldn't be changed** (e.g., data collection, polls)
- **Linear workflows** where going back would break the logic
- **Faster completion** by removing the temptation to second-guess answers
- **Simplified UI** with fewer buttons and options

## Initialization and Configuration

A `Questionaire` is created using `NewBuilder`. You can then chain setter methods to configure it.

```go
// Forward declaration of manager
var manager *questionaire.Manager 
var chatID int64 // Typically from update.Message.Chat.ID
var b *bot.Bot // Your bot instance
ctx := context.Background() // Your context

// Initialize the manager (typically once when your bot starts)
// manager = questionaire.NewManager()

q := questionaire.NewBuilder(chatID, manager).
    SetContext(ctx). // Optional: Set a context for the questionnaire
    SetInitialData(map[string]interface{}{"source": "campaign_xyz"}). // Optional: Pre-fill some data
    SetOnDoneHandler(onSurveyDone). // Handler for when all questions are answered
    SetOnCancelHandler(onSurveyCancel) // Handler for when the user cancels
```

**Setter Methods:**

*   `(*Questionaire) SetManager(m *Manager) *Questionaire`: Sets or updates the manager for this questionnaire.
*   `(*Questionaire) SetContext(ctx context.Context) *Questionaire`: Sets a parent context for the questionnaire. This context will be merged with the context passed to handlers.
*   `(*Questionaire) SetInitialData(data map[string]interface{}) *Questionaire`: Provides initial data that will be included in the final answers map.
*   `(*Questionaire) SetOnDoneHandler(handler onDoneHandlerFunc) *Questionaire`: Sets the callback function to be executed when the questionnaire is successfully completed.
    *   Signature: `func(ctx context.Context, b *bot.Bot, chatID any, answers map[string]interface{}) error`
*   `(*Questionaire) SetOnCancelHandler(handler func()) *Questionaire`: Sets the callback function for when the questionnaire is explicitly cancelled by the user (e.g., via a "Cancel" button).

## Adding Questions

You can add different types of questions to the `Questionaire`.

*   **`(*Questionaire) AddQuestion(key string, text string, choices [][]button.Button, validateFunc func(answer string) error) *Questionaire`**
    *   Adds a question that expects a single answer.
    *   If `choices` is `nil` or empty, it's a `QuestionFormatText` question.
    *   If `choices` are provided, it's a `QuestionFormatRadio` question.
    *   `key`: A unique string to identify this question's answer in the results.
    *   `text`: The question text displayed to the user.
    *   `choices`: A 2D slice of `tgui/button.Button` for radio options. Each inner slice represents a row of buttons. For text questions, pass `nil`.
    *   `validateFunc`: An optional function to validate the user's answer. It should return `nil` if valid, or an `error` if invalid (the error message will be shown to the user).

*   **`(*Questionaire) AddMultipleAnswerQuestion(key string, text string, choices [][]button.Button, validateFunc func(answer string) error) *Questionaire`**
    *   Adds a `QuestionFormatCheck` question (checkbox style).
    *   `choices` are mandatory for this type.
    *   The `validateFunc` here would typically validate individual selections if needed, though often validation for checkboxes is about the overall set of choices (handled after completion).

**Creating Choices with ButtonGrid (Recommended):**

```go
import "github.com/jkevinp/tgui/button"

// For radio questions - clean single choices
radioChoices := button.NewBuilder().
    SingleChoiceWithData("Option A", "opt_a").
    SingleChoiceWithData("Option B", "opt_b").
    Build()

// For checkbox questions - flexible grid layout
checkboxChoices := button.NewBuilder().
    Row().
    ChoiceWithData("Choice X", "choice_x").
    ChoiceWithData("Choice Y", "choice_y").
    Row().
    ChoiceWithData("Choice Z", "choice_z").
    Build()

// Quick helpers for common patterns
yesNoChoices := button.QuickChoices("Yes", "No")
pairedChoices := button.QuickPairedChoices("Option 1", "Option 2", "Option 3", "Option 4")
```

**Legacy Manual Creation (Not Recommended):**

```go
// Old way - manual button creation (harder to maintain)
radioChoices := [][]button.Button{
    {button.Button{Text: "Option A", CallbackData: "opt_a"}}, // Row 1
    {button.Button{Text: "Option B", CallbackData: "opt_b"}}, // Row 2
}
```

**Note:** The `CallbackData` for buttons should be unique for each choice within a question.

**Example of `validateFunc`:**

```go
func validateNonEmpty(answer string) error {
    if strings.TrimSpace(answer) == "" {
        return fmt.Errorf("Answer cannot be empty.")
    }
    return nil
}

func validateNumber(answer string) error {
    if _, err := strconv.Atoi(answer); err != nil {
        return fmt.Errorf("Please enter a valid number.")
    }
    return nil
}
```

## Using the `Manager`

The `Manager` is crucial for handling text-based answers from users.

1.  **Initialization:**
    ```go
    qsManager := questionaire.NewManager()
    ```
    This is typically done once when your bot application starts.

2.  **Handling Messages:**
    You **must** register the `Manager.HandleMessage` method with your `go-telegram/bot` instance to process incoming text messages. This allows the manager to route text replies to the correct active `Questionaire`.
    ```go
    // In your bot setup, after creating the bot instance `b` and `qsManager`
    b.RegisterHandler(bot.HandlerTypeMessageText, "", bot.MatchTypePrefix, qsManager.HandleMessage)
    // Or, if you have a more complex routing system:
    // b.RegisterHandler(bot.HandlerTypeMessageText, "my_text_input_prefix", bot.MatchTypePrefix, qsManager.HandleMessage)
    ```
    When a user sends a text message, and an active questionnaire exists for their `chatID` in the `manager`, `HandleMessage` will call the `Answer()` method of that `Questionaire`.

The `Questionaire.Show()` method, if a manager was provided during `NewBuilder` or via `SetManager`, will automatically add the `Questionaire` instance to the manager. The manager will automatically remove the `Questionaire` instance after the `onDoneHandler` completes successfully or if the `onCancelHandler` is triggered through a managed cancel button.

## Starting and Running the Questionnaire

Once configured, start the questionnaire:

```go
q.Show(ctx, b, chatID)
```

This will send the first question to the user.
*   For text questions, the user replies with a text message, which is caught by `Manager.HandleMessage`.
*   For radio/checkbox questions, the `Questionaire` internally uses `tgui/keyboard/inline` to display buttons. User clicks are handled by internal callback handlers within the `Questionaire`, which then call its `Answer()` method.
*   If an answer is invalid (based on your `validateFunc`), the error is shown, and the question is re-asked.
*   The process continues until all questions are answered or the questionnaire is cancelled.

## Getting Answers

The `onDoneHandler` receives the collected answers:

```go
func onSurveyDone(ctx context.Context, b *bot.Bot, chatID any, answers map[string]interface{}) error {
    // Process answers
    // For QuestionFormatCheck, the answer will be a []string of selected CallbackData
    // For QuestionFormatText/Radio, it will be a string.
    fmt.Printf("Survey for chat %v completed. Answers: %+v\n", chatID, answers)
    b.SendMessage(ctx, &bot.SendMessageParams{
        ChatID: chatID,
        Text:   "Thanks for completing the survey!",
    })
    return nil
}
```
You can also call `q.GetAnswers() map[string]interface{}` on a `Questionaire` instance at any time to get the current state of answers, though it's most relevant in the `onDoneHandler`.

## Interaction Flow (Mermaid Diagram)

```mermaid
sequenceDiagram
    participant User
    participant BotLib ("go-telegram/bot")
    participant AppHandler (Your Bot Code)
    participant QManager ("questionaire.Manager")
    participant QInstance ("questionaire.Questionaire")

    User->>BotLib: Sends /start_survey
    BotLib->>AppHandler: Triggers command handler (e.g., /survey)

    AppHandler->>QInstance: q = questionaire.NewBuilder(chatID, qm)
    AppHandler->>QInstance: q.SetOnDoneHandler(onDoneFunc)
    AppHandler->>QInstance: q.SetOnCancelHandler(onCancelFunc)
    AppHandler->>QInstance: q.AddQuestion("name", "What's your name?", nil, nameValidator)
    AppHandler->>QInstance: q.AddMultipleAnswerQuestion("topics", "Select topics:", topicChoices, nil)
    AppHandler->>QInstance: q.Show(ctx, bot, chatID)

    QInstance->>QManager: (If q.manager is not nil) q.manager.Add(chatID, q)
    QInstance->>BotLib: SendMessage(question 1 text & buttons if any)
    BotLib->>User: Displays question 1

    alt User types text answer (for QuestionFormatText)
        User->>BotLib: Sends text reply
        BotLib->>AppHandler: Triggers general message handler (which should call QManager.HandleMessage)
        AppHandler->>QManager: qm.HandleMessage(ctx, bot, update)
        QManager->>QInstance: q = qm.Get(chatID)
        QManager->>QInstance: isDone = q.Answer(update.Message.Text, bot, chatID)
    else User clicks an inline button (for QuestionFormatRadio/Check)
        User->>BotLib: Clicks inline button
        BotLib->>QInstance: (Routes callback to QInstance's internal handler like onInlineKeyboardSelect, onDoneChoosing)
        QInstance->>QInstance: Internal handler calls isDone = q.Answer(callbackData, bot, chatID)
    end

    QInstance->>QInstance: (Inside Answer) Validates input
    alt Validation Fails
        QInstance->>BotLib: SendMessage(error message to user)
        QInstance->>BotLib: SendMessage(current question again to re-prompt)
        BotLib->>User: Shows error & re-asks question
    else Validation Succeeds
        alt More questions remain OR current question was checkbox type and not "cmd_done"
            QInstance->>QInstance: (If checkbox & not "cmd_done") Updates selected choices
            QInstance->>QInstance: (If not checkbox OR checkbox & "cmd_done") Advances currentQuestionIndex
            QInstance->>BotLib: SendMessage(next question OR current question with updated checkbox UI)
            BotLib->>User: Displays next/updated question
        else All questions answered (isDone = true)
            QInstance->>QInstance: (If called from QManager.HandleMessage) q.Done(ctx, bot, update)
            QInstance->>AppHandler: (Inside q.Done) Calls user's onDoneHandler(ctx, bot, chatID, answers)
            AppHandler->>BotLib: (Typically) SendMessage(confirmation/summary to user)
            QManager->>QManager: (If called from QManager.HandleMessage) qm.Remove(chatID)
        end
    end

    alt User cancels (e.g. clicks cancel button added by q.Show)
        User->>BotLib: Clicks cancel button
        BotLib->>QInstance: (Routes callback to QInstance's onCancel method)
        QInstance->>AppHandler: Calls user's onCancelHandler()
        QInstance->>BotLib: DeleteMessages (clears questionnaire messages)
        QManager->>QManager: (The onCancel handler in Questionaire should ensure q.manager.Remove(chatID) is called if a manager exists)
    end