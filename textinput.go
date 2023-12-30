package main

// A simple program demonstrating the text input component from the Bubbles
// component library.

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// func main() {
// 	p := tea.NewProgram(initialModel())
// 	if _, err := p.Run(); err != nil {
// 		log.Fatal(err)
// 	}
// }

type (
	errMsg error
)

type TextModel struct {
	textInput textinput.Model
	err       error
}

func NewTextModel() TextModel {
	ti := textinput.New()
	// ti.Placeholder = "Pikachu"
	ti.Focus()
	ti.Prompt = ":"
	ti.CharLimit = 156
	ti.Width = 20

	return TextModel{
		textInput: ti,
		err:       nil,
	}
}

func (m TextModel) InitText() tea.Cmd {
	return textinput.Blink
}

func (m TextModel) UpdateText(msg tea.Msg) (TextModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "ctrl+c", "q":
			return m, tea.Quit
		}

	// We handle errors just like any other message
	case errMsg:
		m.err = msg
		return m, nil
	}

	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m TextModel) ViewText(textMode int) string {
	pad := strings.Repeat(" ", padding)
	if textMode == INPUT || textMode == ERROR {
		return pad + fmt.Sprintf(
			m.textInput.View(),
		)
	}
	//TODO: help text
	return ""
}
