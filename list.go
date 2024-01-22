package sptui

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mattn/go-runewidth"
)

const (
	listWidth  = 30
	listHeight = 16
)

type item string

func (i item) FilterValue() string { return "" }

type itemDelegate struct{}

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	str := CleanString(string(i))

	var fn func(strs ...string) string
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render(WrapText("> "+strings.Join(s, " "), listWidth, 2))
		}
	} else {
		fn = func(s ...string) string {
			return itemStyle.Render(PadOrTruncate(strings.Join(s, " "), listWidth))
		}
	}

	fmt.Fprint(w, fn(str))
}

type ListModel struct {
	list     list.Model
	choice   string
	Fetching bool
}

func (m ListModel) InitList() tea.Cmd {
	return nil
}

func (m ListModel) UpdateList(msg tea.Msg, depth int) (ListModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		return m, nil

	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "esc":
			return m, UpdateDepthCmd(-1)

		case "enter":
			i, ok := m.list.SelectedItem().(item)
			if ok {
				m.choice = string(i)
			}

			return m, UpdateDepthCmd(1)
		}
	}

	var cmds []tea.Cmd
	var cmd tea.Cmd

	m.list, cmd = m.list.Update(msg)
	cmds = append(cmds, cmd)

	if len(m.list.Items())-m.list.Index() < 5 {
		cmds = append(cmds, LoadMoreCmd())
	}

	return m, tea.Batch(cmds...)
}

type UpdateDepthMsg struct{ delta int }
type LoadMoreMsg struct{}

func UpdateDepthCmd(d int) tea.Cmd {
	return func() tea.Msg {
		return UpdateDepthMsg{delta: d}
	}
}

func LoadMoreCmd() tea.Cmd {
	return func() tea.Msg {
		return LoadMoreMsg{}
	}
}

func (m ListModel) View(depth int) string {
	return "\n" + m.list.View()
}

func NewListModel(items []list.Item, opts ...ListModelOpt) ListModel {

	l := list.New(items, itemDelegate{}, listWidth, listHeight)

	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle
	l.DisableQuitKeybindings()
	l.SetShowHelp(false)

	m := ListModel{list: l}

	for _, opt := range opts {
		opt(&m)
	}
	return m
}

type ListModelOpt func(*ListModel)

func WithTitle(title string) ListModelOpt {
	return func(m *ListModel) {
		m.list.SetShowTitle(true)
		title = CleanString(title)
		if runewidth.StringWidth(title) <= listWidth {
			m.list.Title = title
		} else {
			m.list.Title = WrapText(title, listWidth, 10)
		}
	}
}
