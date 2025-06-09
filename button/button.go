package button

import (
	"strings"

	"github.com/jkevinp/tgui/keyboard/inline"
)

type Button struct {
	Text         string
	CallbackData string
	OnClick      inline.OnSelect
}

/*
New creates a new Button with the specified text, callback data, and onClick handler.
*/
func New(text string, callbackData string, onClick inline.OnSelect) Button {
	return Button{
		Text:         text,
		CallbackData: callbackData,
		OnClick:      onClick,
	}
}

type ButtonGrid struct {
	Buttons [][]Button
}

/*
NewBuilder creates and returns a new ButtonGrid builder.
*/
func NewBuilder() *ButtonGrid {
	return &ButtonGrid{
		Buttons: make([][]Button, 0),
	}
}

/*
Row adds a new row to the ButtonGrid and returns the builder for chaining.
*/
func (bg *ButtonGrid) Row() *ButtonGrid {
	bg.Buttons = append(bg.Buttons, make([]Button, 0))

	return bg
}

/*
Add appends one or more Buttons to the most recent row in the ButtonGrid.
*/
func (bg *ButtonGrid) Add(btn ...Button) *ButtonGrid {
	bg.Buttons[len(bg.Buttons)-1] = append(bg.Buttons[len(bg.Buttons)-1], btn...)

	return bg
}

/*
Build returns the constructed 2D slice of Buttons from the ButtonGrid.
*/
func (bg *ButtonGrid) Build() [][]Button {
	return bg.Buttons
}

/*
Convenience methods for questionnaire choices
*/

// Choice creates a button with auto-generated callback data and adds it to the current row
func (bg *ButtonGrid) Choice(text string) *ButtonGrid {
	callbackData := strings.ToLower(strings.ReplaceAll(text, " ", "_"))
	callbackData = strings.ReplaceAll(callbackData, "-", "_")
	return bg.Add(Button{Text: text, CallbackData: callbackData})
}

// ChoiceWithData creates a button with custom callback data and adds it to the current row
func (bg *ButtonGrid) ChoiceWithData(text, callbackData string) *ButtonGrid {
	return bg.Add(Button{Text: text, CallbackData: callbackData})
}

// SingleChoice adds a new row with one choice (perfect for radio buttons)
func (bg *ButtonGrid) SingleChoice(text string) *ButtonGrid {
	return bg.Row().Choice(text)
}

// SingleChoiceWithData adds a new row with one choice using custom callback data
func (bg *ButtonGrid) SingleChoiceWithData(text, callbackData string) *ButtonGrid {
	return bg.Row().ChoiceWithData(text, callbackData)
}

/*
Static helper functions for quick choice creation
*/

// QuickChoices creates choices for simple questionnaire scenarios
func QuickChoices(texts ...string) [][]Button {
	builder := NewBuilder()
	for _, text := range texts {
		builder.SingleChoice(text)
	}
	return builder.Build()
}

// QuickPairedChoices creates choices in pairs (2 per row)
func QuickPairedChoices(texts ...string) [][]Button {
	builder := NewBuilder()
	for i := 0; i < len(texts); i += 2 {
		builder.Row().Choice(texts[i])
		if i+1 < len(texts) {
			builder.Choice(texts[i+1])
		}
	}
	return builder.Build()
}

// QuickChoicesWithData creates choices with custom callback data
func QuickChoicesWithData(pairs ...string) [][]Button {
	if len(pairs)%2 != 0 {
		panic("QuickChoicesWithData requires pairs of text and callback data")
	}

	builder := NewBuilder()
	for i := 0; i < len(pairs); i += 2 {
		text := pairs[i]
		callbackData := pairs[i+1]
		builder.SingleChoiceWithData(text, callbackData)
	}
	return builder.Build()
}
