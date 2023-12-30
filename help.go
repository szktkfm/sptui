package main

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
)

type KeyMap struct {
	Play  key.Binding
	Pause key.Binding
	Next  key.Binding
	Prev  key.Binding
	Help  key.Binding
}

type HelpModel struct {
	help   help.Model
	KeyMap KeyMap
}

func NewHelp() HelpModel {
	help := help.New()
	help.Width = 40
	help.ShortSeparator = " "
	return HelpModel{
		help: help,
		KeyMap: KeyMap{
			Play: key.NewBinding(
				key.WithKeys(""),
				key.WithHelp(":play", ""),
			),
			Pause: key.NewBinding(
				key.WithKeys(""),
				key.WithHelp(":pause", ""),
			),
			Next: key.NewBinding(
				key.WithKeys(""),
				key.WithHelp(":next", ""),
			),
			Prev: key.NewBinding(
				key.WithKeys(""),
				key.WithHelp(":prev", ""),
			),
			Help: key.NewBinding(
				key.WithKeys(""),
				key.WithHelp(":help", ""),
			),
			//TODO: add more keybindings
		},
	}

}

func (m HelpModel) ShortHelp() []key.Binding {
	return []key.Binding{
		//TODO: help view
		// m.KeyMap.Help,
		m.KeyMap.Play,
		m.KeyMap.Pause,
		m.KeyMap.Next,
		m.KeyMap.Prev,
	}
}

func (m HelpModel) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{
			m.KeyMap.Play,
			m.KeyMap.Pause,
			m.KeyMap.Next,
			m.KeyMap.Prev,
			m.KeyMap.Help,
		},
	}
}

func (m HelpModel) View() string {
	return m.help.View(m)
}

// DefaultKeyMap returns a default set of keybindings.
func DefaultKeyMap() list.KeyMap {
	return list.KeyMap{
		// Browsing.
		CursorUp: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		CursorDown: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		PrevPage: key.NewBinding(
			key.WithKeys("left", "h", "pgup", "b", "u"),
			key.WithHelp("←/h/pgup", "prev page"),
		),
		NextPage: key.NewBinding(
			key.WithKeys("right", "l", "pgdown", "f", "d"),
			key.WithHelp("→/l/pgdn", "next page"),
		),
		GoToStart: key.NewBinding(
			key.WithKeys("home", "g"),
			key.WithHelp("g/home", "go to start"),
		),
		GoToEnd: key.NewBinding(
			key.WithKeys("end", "G"),
			key.WithHelp("G/end", "go to end"),
		),
		Filter: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "filter"),
		),
		ClearFilter: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "clear filter"),
		),

		// Filtering.
		CancelWhileFiltering: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "cancel"),
		),
		AcceptWhileFiltering: key.NewBinding(
			key.WithKeys("enter", "tab", "shift+tab", "ctrl+k", "up", "ctrl+j", "down"),
			key.WithHelp("enter", "apply filter"),
		),

		// Toggle help.
		// ShowFullHelp: key.NewBinding(
		// 	key.WithKeys("?"),
		// 	key.WithHelp("?", "more"),
		// ),
		// CloseFullHelp: key.NewBinding(
		// 	key.WithKeys("?"),
		// 	key.WithHelp("?", "close help"),
		// ),

		// Quitting.
		Quit: key.NewBinding(
			key.WithKeys("q", "esc"),
			key.WithHelp("q", "quit"),
		),
		ForceQuit: key.NewBinding(key.WithKeys("ctrl+c")),
	}
}
