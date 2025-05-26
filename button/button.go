package button

import "github.com/jkevinp/tgui/keyboard/inline"

type Button struct {
	Text         string
	CallbackData string
	OnClick      inline.OnSelect
}

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

func NewBuilder() *ButtonGrid {
	return &ButtonGrid{
		Buttons: make([][]Button, 0),
	}
}

func (bg *ButtonGrid) Row() *ButtonGrid {
	bg.Buttons = append(bg.Buttons, make([]Button, 0))

	return bg
}

func (bg *ButtonGrid) Add(btn Button) *ButtonGrid {
	bg.Buttons[len(bg.Buttons)-1] = append(bg.Buttons[len(bg.Buttons)-1], btn)

	return bg
}

func (bg *ButtonGrid) Build() [][]Button {
	return bg.Buttons
}
