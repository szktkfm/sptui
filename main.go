package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	tabs := []string{"Playlist", "Album", "Podcast"}
	listModels := []ListModel{
		NewListModel([]list.Item{item(loading)}),
		NewListModel([]list.Item{item(loading)}),
		NewListModel([]list.Item{item(loading)}),
	}

	m := tabModel{
		tabs:        tabs,
		tabContents: listModels,
		depth:       TOP,
		textInput:   NewTextModel(),
		help:        NewHelp(),
	}

	if _, err := tea.NewProgram(m, tea.WithoutSignalHandler()).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
