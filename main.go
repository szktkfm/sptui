package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/uuid"
	"github.com/zmb3/spotify/v2"
)

// Tab
const (
	PLAYLIST = iota
	ALBUM
	PODCAST
)

// Screen Mode
const (
	TOP = iota
	TRACKLIST
)

// Text Input Mode
const (
	NONE = iota
	INPUT
	ERROR
)

var (
	loading           = PadOrTruncate("Loading...", listWidth)
	inactiveTabBorder = tabBorderWithBottom("┴", "─", "┴")
	activeTabBorder   = tabBorderWithBottom("┘", " ", "└")
	docStyle          = lipgloss.NewStyle().Padding(1, 2, 1, 2)
	highlightColor    = lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"}
	inactiveTabStyle  = lipgloss.NewStyle().Border(inactiveTabBorder, true).
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
	errStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("216"))
	helpStyle = list.DefaultStyles().HelpStyle.PaddingLeft(4)
)

type model struct {
	tabs        []string
	activeTab   int
	TabContents []ListModel
	TrackList   ListModel

	Progress BarModel

	TextInput TextModel
	textMode  int

	help HelpModel

	Depth int

	Client     *spotify.Client
	Authorized bool

	Albums        *spotify.SavedAlbumPage
	Playlists     *spotify.SimplePlaylistPage
	Shows         *spotify.SavedShowPage
	SelectedAlbum *spotify.FullAlbum
	//TODO
	Episodes         []spotify.EpisodePage
	SelectedPlaylist *spotify.FullPlaylist

	CurrentlyPlaying *spotify.CurrentlyPlaying
	// PrevTrackID      spotify.ID
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		loginCmd(),
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	if !m.Authorized {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch keypress := msg.String(); keypress {
			case "ctrl+c", "q":
				return m, tea.Quit
			}

		case AuthMsg:
			//Clear screen
			fmt.Print("\033[H\033[2J")
			m.Authorized = true
			m.Client = msg.client
			return m, tea.Batch(
				FetchAlbumsCmd(m.Client),
				GetCurrentlyPlayingTrackCmd(m.Client),
				FetchPlaylistsCmd(m.Client),
				FetchShowsCmd(m.Client),
			)
		default:
			return m, nil
		}
	}

	if m.textMode == INPUT {
		m.TextInput, _ = m.TextInput.UpdateText(msg)
	}

	newProgress, cmd := m.Progress.UpdateBar(msg, m.Client)
	m.Progress = newProgress
	if cmd != nil {
		return m, cmd
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case ":":
			if m.textMode != NONE {
				return m, nil
			}

			m.textMode = INPUT
			m.TextInput.textInput.Prompt = ":"
			return m, nil

		case "enter":
			if m.textMode == ERROR {
				m.textMode = NONE
				m.TextInput = NewTextModel()
				return m, nil
			}

			if m.textMode == INPUT {
				return execTxtCommand(m)
			}

			if m.Depth == TRACKLIST {
				return playTrack(m)
			}
		}

	case CurrentlyPlayingMsg:
		if msg.Track.Item == nil {
			m.Progress = BarModel{}
			return m, nil
		}
		// m.PrevTrackID = msg.Track.Item.SimpleTrack.ID
		m.CurrentlyPlaying = msg.Track

		tickID := uuid.New().String()
		m.Progress = NewBarModel(BarConfig{
			TickID:     tickID,
			Percent:    float64(msg.Track.Progress) / float64(msg.Track.Item.Duration),
			IsPlaying:  msg.Track.Playing,
			DeltaDur:   float64(1000) / float64(msg.Track.Item.Duration),
			TrackTitle: msg.Track.Item.Name + " (" + msg.Track.Item.Artists[0].Name + ")",
		})

		return m, tea.Batch(tickCmd(tickID),
			animTextTickCmd(tickID, 2000*time.Millisecond))

	case PlaybackMsg:
		//TODO: Fix
		time.Sleep(500 * time.Millisecond)
		return m, GetCurrentlyPlayingTrackCmd(m.Client)

	case ErrMsg:
		m.TextInput.textInput.Prompt = "E: "
		m.textMode = ERROR
		m.TextInput.textInput.SetValue(msg.Err.Error())
		return m, nil
	}

	if m.textMode == INPUT {
		return m, nil
	}

	if m.Depth > TOP {
		return listUpdate(m, msg)
	} else {
		return tabUpdate(msg, m)
	}

}

func execTxtCommand(m model) (tea.Model, tea.Cmd) {
	txtCmd := m.TextInput.textInput.Value()
	m.textMode = NONE
	m.TextInput = NewTextModel()

	switch txtCmd {
	case "play":
		if m.CurrentlyPlaying == nil {
			return m, nil
		}
		if m.Progress.IsPlaying {
			return m, nil
		}
		return m, StartPlaybackCmd(m.Client,
			&spotify.PlayOptions{
				URIs:       []spotify.URI{m.CurrentlyPlaying.Item.URI},
				PositionMs: m.CurrentlyPlaying.Progress,
			},
		)
	case "pause":
		return m, PausePlaybackCmd(m.Client)
	case "next":
		return m, NextPlaybackCmd(m.Client)
	case "prev":
		return m, PreviousPlaybackCmd(m.Client)

	default:
		return m, nil
	}
}

func playTrack(m model) (tea.Model, tea.Cmd) {
	selected := m.TrackList.list.Index()
	switch m.activeTab {
	case PLAYLIST:
		return m, StartPlaybackCmd(m.Client,
			&spotify.PlayOptions{
				PlaybackContext: &m.SelectedPlaylist.URI,
				PlaybackOffset: &spotify.PlaybackOffset{
					Position: &selected,
				},
			},
		)

	case ALBUM:
		return m, StartPlaybackCmd(m.Client,
			&spotify.PlayOptions{
				PlaybackContext: &m.SelectedAlbum.URI,
				PlaybackOffset: &spotify.PlaybackOffset{
					Position: &selected,
				},
			},
		)

	case PODCAST:
		return m, StartPlaybackCmd(m.Client,
			&spotify.PlayOptions{
				URIs: []spotify.URI{m.Episodes[selected].URI},
			},
		)
	default:
		return m, nil
	}
}

func listUpdate(m model, msg tea.Msg) (tea.Model, tea.Cmd) {

	switch msg := msg.(type) {
	case AlbumDetailMsg:
		m.SelectedAlbum = msg.Album
		m.TrackList = NewListModel(
			albumTracksToItemList(msg.Album.Tracks.Tracks),
			WithTitle(msg.Album.Name+" ("+msg.Album.Artists[0].Name+")"),
		)

	case ShowDetailMsg:
		m.Episodes = msg.Show.Episodes.Episodes
		m.TrackList = NewListModel(episodesToItemList(m.Episodes),
			WithTitle(msg.Show.Name),
		)

	case PlaylistDetailMsg:
		m.SelectedPlaylist = msg.Playlist
		m.TrackList = NewListModel(playlistTracksToItemList(msg.Playlist.Tracks.Tracks),
			WithTitle(msg.Playlist.Name),
		)

	case UpdateDepthMsg:
		if msg.delta > 0 {
			m.Depth = min(m.Depth+msg.delta, TRACKLIST)
		} else {
			m.Depth = max(m.Depth+msg.delta, TOP)
		}
	}

	newListModel, cmd := m.TrackList.UpdateList(msg, m.Depth)
	m.TrackList = newListModel
	return m, cmd
}

func tabUpdate(msg tea.Msg, m model) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "l", "n", "tab", "right":
			m.activeTab = min(m.activeTab+1, len(m.tabs)-1)
			return m, nil
		case "h", "p", "shift+tab", "left":
			m.activeTab = max(m.activeTab-1, 0)
			return m, nil
		case "j", "down", "k", "up":
			var newListModel ListModel
			newListModel, cmd = m.TabContents[m.activeTab].UpdateList(msg, m.Depth)
			m.TabContents[m.activeTab] = newListModel

		case "enter", " ":
			if m.Playlists == nil || m.Albums == nil || m.Shows == nil {
				return m, nil
			}
			return getTracks(m)
		}

	case AlbumMsg:
		m.Albums = msg.Albums
		m.TabContents[ALBUM] = NewListModel(albumToItemList(msg.Albums))

	case PlaylistMsg:
		m.Playlists = msg.Playlists
		m.TabContents[PLAYLIST] = NewListModel(playlistsToItemList(msg.Playlists))

	case ShowMsg:
		m.Shows = msg.Shows
		m.TabContents[PODCAST] = NewListModel(showsToItemList(msg.Shows))

	}
	return m, cmd
}

func getTracks(m model) (tea.Model, tea.Cmd) {

	m.Depth = TRACKLIST
	m.TrackList = NewListModel([]list.Item{item(loading)})

	selected := m.TabContents[m.activeTab].list.Index()
	switch m.activeTab {
	case PLAYLIST:
		return m, GetPlaylistCmd(m.Client, m.Playlists.Playlists[selected].ID)
	case ALBUM:
		return m, GetAlbumCmd(m.Client, m.Albums.Albums[selected].ID)
	case PODCAST:
		return m, GetShowCmd(m.Client, m.Shows.Shows[selected].ID)
	default:
		return m, nil
	}
}

func playlistTracksToItemList(tracks []spotify.PlaylistTrack) []list.Item {
	var itemList []list.Item
	for _, t := range tracks {
		itemList = append(itemList, item(t.Track.Name))
	}
	return itemList
}

func episodesToItemList(episodes []spotify.EpisodePage) []list.Item {
	var itemList []list.Item
	for _, e := range episodes {
		itemList = append(itemList, item(e.Name))
	}
	return itemList
}

func albumTracksToItemList(tracks []spotify.SimpleTrack) []list.Item {
	var itemList []list.Item
	for _, t := range tracks {
		itemList = append(itemList, item(t.Name))
	}
	return itemList
}

func albumToItemList(albums *spotify.SavedAlbumPage) []list.Item {
	// TODO:added_atでソート
	var itemList []list.Item
	for _, a := range albums.Albums {
		itemList = append(itemList, item(a.Name))
	}
	return itemList
}

func showsToItemList(shows *spotify.SavedShowPage) []list.Item {
	var itemList []list.Item
	for _, s := range shows.Shows {
		itemList = append(itemList, item(s.Name))
	}
	return itemList
}

func playlistsToItemList(playlist *spotify.SimplePlaylistPage) []list.Item {
	var itemList []list.Item
	for _, p := range playlist.Playlists {
		itemList = append(itemList, item(p.Name))
	}
	return itemList
}

func tabBorderWithBottom(left, middle, right string) lipgloss.Border {
	border := lipgloss.RoundedBorder()
	border.BottomLeft = left
	border.Bottom = middle
	border.BottomRight = right
	return border
}

func paddingTabBorder() string {
	style := activeTabStyle.Copy()
	border, _, _, _, _ := style.GetBorder()
	border.Bottom = "─"
	border.BottomLeft = "─"
	border.BottomRight = "╮"
	border.Top = ""
	border.TopLeft = ""
	border.TopRight = ""
	border.MiddleLeft = ""
	border.MiddleRight = ""
	border.Left = ""
	border.Right = ""

	style = style.Border(border)
	return style.Render(strings.Repeat(" ", 8))
}

func (m model) View() string {

	if !m.Authorized {
		return ""
	}

	var view string
	if m.Depth > TOP {
		view += tracksView(m)
	} else {
		view += tabView(m)
	}

	if m.CurrentlyPlaying != nil {
		view += progressView(m)
	}

	view += textInputView(m)

	return view
}

func progressView(m model) string {
	return "\n" + m.Progress.ViewBar()
}

func textInputView(m model) string {
	switch m.textMode {
	case INPUT:
		return "\n" + m.TextInput.ViewText(m.textMode)
	case ERROR:
		return "\n" + errStyle.Render(m.TextInput.ViewText(m.textMode))
	case NONE:
		return helpStyle.Render(m.help.View())
	default:
		return ""
	}
}

func tracksView(m model) string {
	doc := strings.Builder{}
	windowStyleDtl := lipgloss.NewStyle().
		BorderForeground(highlightColor).
		Padding(1, 5).
		Align(lipgloss.Left).
		Border(lipgloss.RoundedBorder())

	doc.WriteString(
		windowStyleDtl.
			Render(m.TrackList.View(m.Depth)))
	return docStyle.Render(doc.String())
}

func tabView(m model) string {
	doc := strings.Builder{}

	var renderedTabs []string

	for i, t := range m.tabs {
		var style lipgloss.Style
		isFirst, isLast, isActive := i == 0, i == len(m.tabs)-1, i == m.activeTab
		if isActive {
			style = activeTabStyle.Copy()
		} else {
			style = inactiveTabStyle.Copy()
		}
		border, _, _, _, _ := style.GetBorder()
		if isFirst && isActive {
			border.BottomLeft = "│"
		} else if isFirst && !isActive {
			border.BottomLeft = "├"
		} else if isLast && isActive {
			border.BottomRight = "└"
		} else if isLast && !isActive {
			border.BottomRight = "┴"
		}
		style = style.Border(border)
		renderedTabs = append(renderedTabs, style.Render(t))
	}

	renderedTabs = append(renderedTabs, paddingTabBorder())
	row := lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...)
	doc.WriteString(row)
	doc.WriteString("\n")
	doc.WriteString(
		windowStyle.
			// Width((lipgloss.Width(row) - windowStyle.GetHorizontalFrameSize())).
			Render(m.TabContents[m.activeTab].View(m.Depth)))

	return docStyle.Render(doc.String())
}

func main() {
	tabs := []string{"Playlist", "Album", "Podcast"}
	listModels := []ListModel{
		NewListModel([]list.Item{item(loading)}),
		NewListModel([]list.Item{item(loading)}),
		NewListModel([]list.Item{item(loading)}),
	}

	m := model{
		tabs:        tabs,
		TabContents: listModels,
		Depth:       TOP,
		TextInput:   NewTextModel(),
		help:        NewHelp(),
	}

	if _, err := tea.NewProgram(m, tea.WithoutSignalHandler()).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
