package sptui

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
	maxWidth = 42
)

var trackTitleStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("212")).Render

type tickMsg struct {
	id   string
	time time.Time
}

type BarModel struct {
	IsPlaying  bool
	percent    float64
	progress   progress.Model
	deltaDur   float64
	tickID     string
	trackTitle string
	titleAnim  AnimTextModel
	animate    bool
}

func (m BarModel) UpdateBar(msg tea.Msg, client *spotify.Client) (BarModel, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.progress.Width = msg.Width - padding*2 - 4
		if m.progress.Width > maxWidth {
			m.progress.Width = maxWidth
		}
		return m, nil

	case tickMsg:
		if m.tickID != msg.id {
			return m, nil
		}
		if m.IsPlaying {
			m.percent += m.deltaDur
			if m.percent > 1.0 {
				m.percent = 1.0
				return m, GetCurrentlyPlayingTrackCmd(client)
			}
		}
		return m, tickCmd(m.tickID)
	}
	if m.animate {
		newTitleAnim, cmd := m.titleAnim.UpdateAnimText(msg)
		m.titleAnim = newTitleAnim
		if cmd != nil {
			return m, cmd
		}
	}
	return m, nil
}

func (m BarModel) ViewBar() string {
	pad := strings.Repeat(" ", padding)
	view := pad + "🎧 "

	if m.animate {
		view += trackTitleStyle(m.titleAnim.ViewAnimText()) + "\n"
	} else {
		view += trackTitleStyle(m.trackTitle) + "\n"
	}
	view += pad + m.progress.ViewAs(m.percent)
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
		progress:   prog,
		percent:    conf.Percent,
		IsPlaying:  conf.IsPlaying,
		deltaDur:   conf.DeltaDur,
		tickID:     conf.TickID,
		trackTitle: conf.TrackTitle,
	}
	if len(conf.TrackTitle) > maxWidth {
		m.titleAnim = NewAnimText(conf.TrackTitle, conf.TickID,
			WithWidth(maxWidth-4))
		m.animate = true
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
