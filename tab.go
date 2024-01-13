package spotui

import (
	"fmt"
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
)

type TabModel struct {
	tabs        []string
	activeTab   int
	tabContents []ListModel
	trackList   ListModel

	progress BarModel

	textInput TextModel
	textMode  int

	help HelpModel

	depth int

	client     *spotify.Client
	authorized bool

	albums        *spotify.SavedAlbumPage
	playlists     *spotify.SimplePlaylistPage
	shows         *spotify.SavedShowPage
	selectedAlbum *spotify.FullAlbum
	//TODO
	episodes         []spotify.EpisodePage
	selectedPlaylist *spotify.FullPlaylist

	currentlyPlaying *spotify.CurrentlyPlaying
}

func (m TabModel) Init() tea.Cmd {
	return tea.Batch(
		loginCmd(),
	)
}

func (m TabModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	if !m.authorized {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch keypress := msg.String(); keypress {
			case "ctrl+c", "q":
				return m, tea.Quit
			}

		case AuthMsg:
			//Clear screen
			fmt.Print("\033[H\033[2J")
			m.authorized = true
			m.client = msg.client
			return m, tea.Batch(
				FetchAlbumsCmd(m.client),
				GetCurrentlyPlayingTrackCmd(m.client),
				FetchPlaylistsCmd(m.client),
				FetchShowsCmd(m.client),
			)
		default:
			return m, nil
		}
	}

	if m.textMode == INPUT {
		m.textInput, _ = m.textInput.UpdateText(msg)
	}

	newProgress, cmd := m.progress.UpdateBar(msg, m.client)
	m.progress = newProgress
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
			m.textInput.textInput.Prompt = ":"
			return m, nil

		case "enter":
			if m.textMode == ERROR {
				m.textMode = NONE
				m.textInput = NewTextModel()
				return m, nil
			}

			if m.textMode == INPUT {
				return execTxtCommand(m)
			}

			if m.depth == TRACKLIST {
				return playTrack(m)
			}
		}

	case CurrentlyPlayingMsg:
		if msg.Track.Item == nil {
			m.progress = BarModel{}
			return m, nil
		}
		m.currentlyPlaying = msg.Track

		tickID := uuid.New().String()
		m.progress = NewBarModel(BarConfig{
			TickID:     tickID,
			Percent:    float64(msg.Track.Progress) / float64(msg.Track.Item.Duration),
			IsPlaying:  msg.Track.Playing,
			DeltaDur:   float64(1000) / float64(msg.Track.Item.Duration),
			TrackTitle: msg.Track.Item.Name + " (" + msg.Track.Item.Artists[0].Name + ")",
		})

		return m, tea.Batch(tickCmd(tickID),
			AnimTextTickCmd(tickID, 2000*time.Millisecond))

	case PlaybackMsg:
		//TODO: Fix
		time.Sleep(500 * time.Millisecond)
		return m, GetCurrentlyPlayingTrackCmd(m.client)

	case ErrMsg:
		m.textInput.textInput.Prompt = "E: "
		m.textMode = ERROR
		m.textInput.textInput.SetValue(msg.Err.Error())
		return m, nil
	}

	if m.textMode == INPUT {
		return m, nil
	}

	if m.depth > TOP {
		return listUpdate(m, msg)
	} else {
		return tabUpdate(msg, m)
	}

}

func execTxtCommand(m TabModel) (tea.Model, tea.Cmd) {
	txtCmd := m.textInput.textInput.Value()
	m.textMode = NONE
	m.textInput = NewTextModel()

	switch txtCmd {
	case "play":
		if m.currentlyPlaying == nil {
			return m, nil
		}
		if m.progress.IsPlaying {
			return m, nil
		}
		return m, StartPlaybackCmd(m.client,
			&spotify.PlayOptions{
				URIs:       []spotify.URI{m.currentlyPlaying.Item.URI},
				PositionMs: m.currentlyPlaying.Progress,
			},
		)
	case "pause":
		return m, PausePlaybackCmd(m.client)
	case "next":
		return m, NextPlaybackCmd(m.client)
	case "prev":
		return m, PreviousPlaybackCmd(m.client)

	default:
		return m, nil
	}
}

func playTrack(m TabModel) (tea.Model, tea.Cmd) {
	selected := m.trackList.list.Index()
	switch m.activeTab {
	case PLAYLIST:
		return m, StartPlaybackCmd(m.client,
			&spotify.PlayOptions{
				PlaybackContext: &m.selectedPlaylist.URI,
				PlaybackOffset: &spotify.PlaybackOffset{
					Position: &selected,
				},
			},
		)

	case ALBUM:
		return m, StartPlaybackCmd(m.client,
			&spotify.PlayOptions{
				PlaybackContext: &m.selectedAlbum.URI,
				PlaybackOffset: &spotify.PlaybackOffset{
					Position: &selected,
				},
			},
		)

	case PODCAST:
		return m, StartPlaybackCmd(m.client,
			&spotify.PlayOptions{
				URIs: []spotify.URI{m.episodes[selected].URI},
			},
		)
	default:
		return m, nil
	}
}

func listUpdate(m TabModel, msg tea.Msg) (tea.Model, tea.Cmd) {

	switch msg := msg.(type) {
	case AlbumDetailMsg:
		m.selectedAlbum = msg.Album
		m.trackList = NewListModel(
			albumTracksToItemList(msg.Album.Tracks.Tracks),
			WithTitle(msg.Album.Name+" ("+msg.Album.Artists[0].Name+")"),
		)

	case ShowDetailMsg:
		m.episodes = msg.Show.Episodes.Episodes
		m.trackList = NewListModel(episodesToItemList(m.episodes),
			WithTitle(msg.Show.Name),
		)

	case PlaylistDetailMsg:
		m.selectedPlaylist = msg.Playlist
		m.trackList = NewListModel(playlistTracksToItemList(msg.Playlist.Tracks.Tracks),
			WithTitle(msg.Playlist.Name),
		)

	case UpdateDepthMsg:
		if msg.delta > 0 {
			m.depth = min(m.depth+msg.delta, TRACKLIST)
		} else {
			m.depth = max(m.depth+msg.delta, TOP)
		}
	}

	newListModel, cmd := m.trackList.UpdateList(msg, m.depth)
	m.trackList = newListModel
	return m, cmd
}

func tabUpdate(msg tea.Msg, m TabModel) (tea.Model, tea.Cmd) {
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
			if m.playlists == nil || m.albums == nil || m.shows == nil {
				return m, nil
			}
			var newListModel ListModel
			newListModel, cmd = m.tabContents[m.activeTab].UpdateList(msg, m.depth)
			m.tabContents[m.activeTab] = newListModel

		case "enter", " ":
			if m.playlists == nil || m.albums == nil || m.shows == nil {
				return m, nil
			}
			return getTracks(m)
		}

	case AlbumMsg:
		if m.albums == nil {
			m.albums = msg.Albums
			m.tabContents[ALBUM] = NewListModel(albumToItemList(msg.Albums))
		} else {
			m.albums.Offset = msg.Albums.Offset
			newAlbums := append(m.albums.Albums, msg.Albums.Albums...)
			m.albums.Albums = newAlbums

			newListModel := NewListModel(albumToItemList(m.albums))
			newListModel.list.Select(m.tabContents[ALBUM].list.Index())
			m.tabContents[ALBUM] = newListModel
		}
		m.tabContents[ALBUM].Fetching = false

	case PlaylistMsg:
		if m.playlists == nil {
			m.playlists = msg.Playlists
			m.tabContents[PLAYLIST] = NewListModel(playlistsToItemList(msg.Playlists))
		} else {
			m.playlists.Offset = msg.Playlists.Offset
			newPlaylists := append(m.playlists.Playlists, msg.Playlists.Playlists...)
			m.playlists.Playlists = newPlaylists

			newListModel := NewListModel(playlistsToItemList(m.playlists))
			newListModel.list.Select(m.tabContents[PLAYLIST].list.Index())
			m.tabContents[PLAYLIST] = newListModel
		}
		m.tabContents[PLAYLIST].Fetching = false

	case ShowMsg:
		if m.shows == nil {
			m.shows = msg.Shows
			m.tabContents[PODCAST] = NewListModel(showsToItemList(msg.Shows))
		} else {
			m.shows.Offset = msg.Shows.Offset
			newShows := append(m.shows.Shows, msg.Shows.Shows...)
			m.shows.Shows = newShows

			newListModel := NewListModel(showsToItemList(m.shows))
			newListModel.list.Select(m.tabContents[PODCAST].list.Index())
			m.tabContents[PODCAST] = newListModel
		}
		m.tabContents[PODCAST].Fetching = false

	case LoadMoreMsg:
		switch m.activeTab {
		case PLAYLIST:
			if m.tabContents[PLAYLIST].Fetching {
				return m, nil
			}
			m.tabContents[PLAYLIST].Fetching = true
			return m, FetchPlaylistsCmd(m.client, spotify.Offset(m.playlists.Offset+len(m.playlists.Playlists)))
		case ALBUM:
			if m.tabContents[ALBUM].Fetching {
				return m, nil
			}
			m.tabContents[ALBUM].Fetching = true
			return m, FetchAlbumsCmd(m.client, spotify.Offset(m.albums.Offset+len(m.albums.Albums)))
		case PODCAST:
			if m.tabContents[PODCAST].Fetching {
				return m, nil
			}
			m.tabContents[PODCAST].Fetching = true
			return m, FetchShowsCmd(m.client, spotify.Offset(m.shows.Offset+len(m.shows.Shows)))

		}
	}
	return m, cmd
}

func getTracks(m TabModel) (tea.Model, tea.Cmd) {

	m.depth = TRACKLIST
	m.trackList = NewListModel([]list.Item{item(loading)})

	selected := m.tabContents[m.activeTab].list.Index()
	switch m.activeTab {
	case PLAYLIST:
		return m, GetPlaylistCmd(m.client, m.playlists.Playlists[selected].ID)
	case ALBUM:
		return m, GetAlbumCmd(m.client, m.albums.Albums[selected].ID)
	case PODCAST:
		return m, GetShowCmd(m.client, m.shows.Shows[selected].ID)
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

func (m TabModel) View() string {

	if !m.authorized {
		return ""
	}

	var view string
	if m.depth > TOP {
		view += tracksView(m)
	} else {
		view += tabView(m)
	}

	if m.currentlyPlaying != nil {
		view += progressView(m)
	}

	view += textInputView(m)

	return view
}

func progressView(m TabModel) string {
	return "\n" + m.progress.ViewBar()
}

func textInputView(m TabModel) string {
	switch m.textMode {
	case INPUT:
		return "\n" + m.textInput.ViewText(m.textMode)
	case ERROR:
		return "\n" + errStyle.Render(m.textInput.ViewText(m.textMode))
	case NONE:
		return helpStyle.Render(m.help.View())
	default:
		return ""
	}
}

func tracksView(m TabModel) string {
	doc := strings.Builder{}
	windowStyleDtl := lipgloss.NewStyle().
		BorderForeground(highlightColor).
		Padding(1, 5).
		Align(lipgloss.Left).
		Border(lipgloss.RoundedBorder())

	doc.WriteString(
		windowStyleDtl.
			Render(m.trackList.View(m.depth)))
	return docStyle.Render(doc.String())
}

func tabView(m TabModel) string {
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
			Render(m.tabContents[m.activeTab].View(m.depth)))

	return docStyle.Render(doc.String())
}

func NewTabModel() TabModel {
	tabs := []string{"Playlist", "Album", "Podcast"}
	listModels := []ListModel{
		NewListModel([]list.Item{item(loading)}),
		NewListModel([]list.Item{item(loading)}),
		NewListModel([]list.Item{item(loading)}),
	}

	return TabModel{
		tabs:        tabs,
		tabContents: listModels,
		depth:       TOP,
		textInput:   NewTextModel(),
		help:        NewHelp(),
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
