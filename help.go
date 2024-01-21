package sptui

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
)

type KeyMap struct {
	Play   key.Binding
	Pause  key.Binding
	Next   key.Binding
	Prev   key.Binding
	Help   key.Binding
	Device key.Binding
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
			Device: key.NewBinding(
				key.WithKeys(""),
				key.WithHelp(":device", ""),
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
		m.KeyMap.Device,
	}
}

func (m HelpModel) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{
			m.KeyMap.Play,
			m.KeyMap.Pause,
			m.KeyMap.Next,
			m.KeyMap.Prev,
			m.KeyMap.Device,
			m.KeyMap.Help,
		},
	}
}

func (m HelpModel) View() string {
	return m.help.View(m)
}
