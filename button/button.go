package button

import "github.com/jkevinp/tgui/keyboard/inline"

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
