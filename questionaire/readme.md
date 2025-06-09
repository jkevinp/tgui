# Questionnaire Component

The `questionaire` component for `tgui` allows you to guide a Telegram bot user through a sequence of questions, collect their answers, and process them. It supports text input, single-choice (radio button style), and multiple-choice (checkbox style) questions.

## Quick Start

**Try the live demo:** You can test the questionnaire component right now by messaging [@TGUIEXbot](https://t.me/TGUIEXbot) and sending `/survey`.

To run the example locally:
1. Navigate to `/home/joketa/tgui/examples/`
2. The `.env` file is already configured with a bot token
3. Run: `go run questionaire_example.go`
4. Send `/survey` to the bot to test the questionnaire functionality

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

**Example of `choices` for `button.Button`:**

```go
import "github.com/jkevinp/tgui/button"

// For a radio or checkbox question
radioChoices := [][]button.Button{
    {button.Button{Text: "Option A", CallbackData: "opt_a"}}, // Row 1
    {button.Button{Text: "Option B", CallbackData: "opt_b"}}, // Row 2
}

checkboxChoices := [][]button.Button{
    {
        button.Button{Text: "Choice X", CallbackData: "choice_x"},
        button.Button{Text: "Choice Y", CallbackData: "choice_y"},
    }, // Row 1
    {button.Button{Text: "Choice Z", CallbackData: "choice_z"}}, // Row 2
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