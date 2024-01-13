# spotui
[![Go Report Card](https://goreportcard.com/badge/github.com/szktkfm/spotui)](https://goreportcard.com/report/github.com/szktkfm/spotui)

<img src="assets/demo.gif" width="500">

## Overview
spotui is a Spotify TUI player, written in Go and leveraging the  [charmbracelet/bubbletea](https://github.com/charmbracelet/bubbletea)  library. 

## Installation
Visit the  [GitHub Releases](https://github.com/szktkfm/spotui/releases) page for spotui and download the appropriate binary for your operating system.

Or, just install it with `go`:
```bash
go install github.com/szktkfm/spotui@latest
```

## Usage
### Connecting to Spotify’s API
To use spotui, you need to connect it to Spotify's API. Follow these steps:

1. [Go to the Spotify Dashboard](https://developer.spotify.com/dashboard).
2. Create a new app to obtain your Client ID and Client Secret.
3. In 'Edit Settings', add http://localhost:21112/callback to the 'Redirect URIs'. Don’t forget to save your changes.
4. Set your Client ID as an environment variable `SPOTIFY_ID`. 
5. Run spotui. You will see an official Spotify URL for authentication.

```bash
# Replace your_client_id with the actual Client ID you obtained from Spotify.
SPOTIFY_ID=your_client_id spotui
```
6. Open the provided URL in a web browser and log in to your Spotify account to grant the necessary permissions.
After granting permission, you might be redirected to a blank page. This is normal and indicates that the authentication process is complete.

Once authenticated, you are ready to use spotui!

### API Token Storage
Once authenticated, your Spotify API token will be stored at `${HOME}/.config/spotui/spotify_token.json`. Ensure this file is kept secure as it contains sensitive information.

### Key Bindings
Here are the key bindings for spotui:

| Key       | Action                           |
|-----------|----------------------------------|
| `h` `j` `k` `l` | Navigate (left, down, up, right) |
| `q`       | Quit spotui                      |
| `:play`   | Play current selection           |
| `:pause`  | Pause playback                   |
| `:next`   | Next track                       |
| `:prev`   | Previous track                   |

