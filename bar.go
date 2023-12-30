package main

import (
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/zmb3/spotify/v2"
)

const (
	padding  = 2
	maxWidth = 44
)

var trackTitleStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("212")).Render

type tickMsg struct {
	id   string
	time time.Time
}

type BarModel struct {
	Percent    float64
	Progress   progress.Model
	IsPlaying  bool
	DeltaDur   float64
	tickID     string
	TrackTitle string
	TitleAnim  AnimTextModel
	animation  bool
}

func (m BarModel) UpdateBar(msg tea.Msg, client *spotify.Client) (BarModel, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.Progress.Width = msg.Width - padding*2 - 4
		if m.Progress.Width > maxWidth {
			m.Progress.Width = maxWidth
		}
		return m, nil

	case tickMsg:
		if m.tickID != msg.id {
			return m, nil
		}
		if m.IsPlaying {
			m.Percent += m.DeltaDur
			if m.Percent > 1.0 {
				m.Percent = 1.0
				return m, GetCurrentlyPlayingTrackCmd(client)
			}
		}
		return m, tickCmd(m.tickID)
	}
	if m.animation {
		newTitleAnim, cmd := m.TitleAnim.UpdateAnimText(msg)
		m.TitleAnim = newTitleAnim
		if cmd != nil {
			return m, cmd
		}
	}
	return m, nil
}

func (m BarModel) ViewBar() string {
	pad := strings.Repeat(" ", padding)
	view := pad + "🎧 "

	if m.animation {
		view += trackTitleStyle(m.TitleAnim.ViewAnimText()) + "\n"
	} else {
		view += trackTitleStyle(m.TrackTitle) + "\n"
	}
	view += pad + m.Progress.ViewAs(m.Percent)
	return view
}

type BarConfig struct {
	TickID     string
	Percent    float64
	IsPlaying  bool
	DeltaDur   float64
	TrackTitle string
}

func NewBarModel(conf BarConfig) BarModel {
	prog := progress.New(
		progress.WithWidth(44),
		progress.WithoutPercentage(),
		progress.WithDefaultScaledGradient(),
	)
	m := BarModel{
		Progress:   prog,
		Percent:    conf.Percent,
		IsPlaying:  conf.IsPlaying,
		DeltaDur:   conf.DeltaDur,
		tickID:     conf.TickID,
		TrackTitle: conf.TrackTitle,
	}
	if len(conf.TrackTitle) > maxWidth {
		m.TitleAnim = NewAnimText(conf.TrackTitle, conf.TickID,
			WithWidth(maxWidth-4))
		m.animation = true
	}
	return m
}

func tickCmd(id string) tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg{
			time: t,
			id:   id,
		}
	})
}
