package sptui

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
)

var (
	docStyle         = lipgloss.NewStyle().Padding(1, 2, 1, 2)
	highlightColor   = lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"}
	inactiveTabStyle = lipgloss.NewStyle().Border(inactiveTabBorder, true).
				BorderForeground(highlightColor).Padding(0, 1)
	activeTabStyle = inactiveTabStyle.Copy().Border(activeTabBorder, true)
	borderStyle    = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, false, false, true).
			BorderForeground(lipgloss.Color("5"))

	windowStyle = lipgloss.NewStyle().
			BorderForeground(highlightColor).
			Padding(1, 5).
			Align(lipgloss.Left).
			Border(lipgloss.NormalBorder()).
			UnsetBorderTop()
	errStyle          = lipgloss.NewStyle().Foreground(lipgloss.Color("216"))
	helpStyle         = list.DefaultStyles().HelpStyle.PaddingLeft(2)
	itemStyle         = lipgloss.NewStyle().PaddingLeft(2)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	paginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
)
