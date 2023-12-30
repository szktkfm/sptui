package main

import (
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type AnimTextModel struct {
	text   string
	offset int
	width  int
	id     string
}

func (m AnimTextModel) UpdateAnimText(msg tea.Msg) (AnimTextModel, tea.Cmd) {
	switch msg.(type) {
	case animTextTickMsg:
		if m.id != msg.(animTextTickMsg).id {
			return m, nil
		}
		m.offset = (m.offset + 1) % len(m.text)
		if m.offset == 0 {
			return m, animTextTickCmd(m.id, 2000*time.Millisecond)
		}
		return m, animTextTickCmd(m.id, 100*time.Millisecond)
	}

	return m, nil
}

func (m AnimTextModel) ViewAnimText() string {
	left := m.offset
	right := (left + m.width) % len(m.text)

	if left < right {
		return m.text[left:right]
	} else {
		return m.text[left:] + m.text[:right]
	}
}

func NewAnimText(t string, id string, opts ...AnimTextModelOpt) AnimTextModel {
	m := AnimTextModel{
		text:   t + strings.Repeat(" ", 4),
		width:  listWidth - 4,
		offset: 0,
		id:     id,
	}

	for _, opt := range opts {
		opt(&m)
	}
	return m
}

func animTextTickCmd(id string, t time.Duration) tea.Cmd {
	return tea.Tick(t, func(t time.Time) tea.Msg {
		return animTextTickMsg{
			time: t,
			id:   id,
		}
	})
}

type animTextTickMsg struct {
	id   string
	time time.Time
}

type AnimTextModelOpt func(*AnimTextModel)

func WithWidth(w int) AnimTextModelOpt {
	return func(m *AnimTextModel) {
		m.width = w
	}
}
