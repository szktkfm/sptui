package main

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/zmb3/spotify/v2"
)

type ErrMsg struct {
	Err error
}

type AlbumMsg struct {
	Albums *spotify.SavedAlbumPage
}

type PlaylistMsg struct {
	Playlists *spotify.SimplePlaylistPage
}

type ShowMsg struct {
	Shows *spotify.SavedShowPage
}

type AlbumDetailMsg struct {
	Album *spotify.FullAlbum
}

type ShowDetailMsg struct {
	Show *spotify.FullShow
}

type PlaylistDetailMsg struct {
	Playlist *spotify.FullPlaylist
}

type CurrentlyPlayingMsg struct {
	Track *spotify.CurrentlyPlaying
}

type PlaybackMsg struct {
}

func FetchAlbumsCmd(client *spotify.Client, opts ...spotify.RequestOption) tea.Cmd {
	return func() tea.Msg {
		albums, err := client.CurrentUsersAlbums(context.Background(), opts...)
		if err != nil {
			return ErrMsg{Err: err}
		}
		return AlbumMsg{Albums: albums}
	}
}

func FetchPlaylistsCmd(client *spotify.Client, opts ...spotify.RequestOption) tea.Cmd {
	return func() tea.Msg {
		playlist, err := client.CurrentUsersPlaylists(context.Background(), opts...)
		if err != nil {
			return ErrMsg{Err: err}
		}
		return PlaylistMsg{Playlists: playlist}
	}
}

func FetchShowsCmd(client *spotify.Client, opts ...spotify.RequestOption) tea.Cmd {
	return func() tea.Msg {
		shows, err := client.CurrentUsersShows(context.Background(), opts...)
		if err != nil {
			return ErrMsg{Err: err}
		}
		return ShowMsg{Shows: shows}
	}
}

func GetAlbumCmd(client *spotify.Client, id spotify.ID) tea.Cmd {
	return func() tea.Msg {
		album, err := client.GetAlbum(context.Background(), id)
		if err != nil {
			return ErrMsg{Err: err}
		}
		return AlbumDetailMsg{Album: album}
	}
}

func GetShowCmd(client *spotify.Client, id spotify.ID) tea.Cmd {
	return func() tea.Msg {
		show, err := client.GetShow(context.Background(), id)
		if err != nil {
			return ErrMsg{Err: err}
		}
		return ShowDetailMsg{Show: show}
	}
}

func GetPlaylistCmd(client *spotify.Client, id spotify.ID) tea.Cmd {
	return func() tea.Msg {
		playlist, err := client.GetPlaylist(context.Background(), id)
		if err != nil {
			return ErrMsg{Err: err}
		}
		return PlaylistDetailMsg{Playlist: playlist}
	}
}

func GetCurrentlyPlayingTrackCmd(client *spotify.Client) tea.Cmd {
	return func() tea.Msg {
		track, err := client.PlayerCurrentlyPlaying(context.Background())
		if err != nil {
			return ErrMsg{Err: err}
		}

		//TODO: podcast support
		return CurrentlyPlayingMsg{Track: track}
	}
}

func StartPlaybackCmd(client *spotify.Client, opts *spotify.PlayOptions) tea.Cmd {
	return func() tea.Msg {
		err := client.PlayOpt(context.Background(),
			opts,
		)
		if err != nil {
			return ErrMsg{Err: err}
		}
		return PlaybackMsg{}
	}
}

func PausePlaybackCmd(client *spotify.Client) tea.Cmd {
	return func() tea.Msg {
		err := client.Pause(context.Background())
		if err != nil {
			return ErrMsg{Err: err}
		}
		return PlaybackMsg{}
	}
}

func NextPlaybackCmd(client *spotify.Client) tea.Cmd {
	return func() tea.Msg {
		err := client.Next(context.Background())
		if err != nil {
			return ErrMsg{Err: err}
		}
		return PlaybackMsg{}
	}
}

func PreviousPlaybackCmd(client *spotify.Client) tea.Cmd {
	return func() tea.Msg {
		err := client.Previous(context.Background())
		if err != nil {
			return ErrMsg{Err: err}
		}
		return PlaybackMsg{}
	}
}
