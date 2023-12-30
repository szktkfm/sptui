package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/pkg/browser"
	spotifyauth "github.com/zmb3/spotify/v2/auth"

	"golang.org/x/oauth2"

	"github.com/zmb3/spotify/v2"
)

const redirectURI = "http://localhost:21112/callback"

var (
	auth = spotifyauth.New(spotifyauth.WithRedirectURL(redirectURI),
		spotifyauth.WithScopes(
			spotifyauth.ScopeUserReadPrivate,
			spotifyauth.ScopeUserModifyPlaybackState,
			spotifyauth.ScopeUserLibraryRead,
			spotifyauth.ScopePlaylistReadCollaborative,
			spotifyauth.ScopePlaylistReadPrivate,
			spotifyauth.ScopeUserReadCurrentlyPlaying,
		))
	tokenCh       = make(chan *oauth2.Token)
	tokenFilePath = ".config/spotui/spotify_token.json"
)

func generateCodeVerifier() (string, error) {
	randomBytes := make([]byte, 32)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(randomBytes), nil
}

func generateCodeChallenge(verifier string) string {
	sha256Sum := sha256.Sum256([]byte(verifier))
	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(sha256Sum[:])
}

func generateState() string {
	randomBytes := make([]byte, 16)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return ""
	}
	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(randomBytes)
}

type TokenExpiredError struct {
	Message string
}

func (e TokenExpiredError) Error() string {
	return e.Message
}

func NewTokenExpiredError() error {
	return TokenExpiredError{Message: "Token is expired"}
}

func isTokenExpiredError(err error) bool {
	_, ok := err.(TokenExpiredError)
	return ok
}

type AuthMsg struct {
	client *spotify.Client
}

func saveOAuthToken(token *oauth2.Token) error {
	homeDir, _ := os.UserHomeDir()
	dir := filepath.Dir(filepath.Join(homeDir, tokenFilePath))
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.MkdirAll(dir, 0700)
	}

	jsonToken, err := json.Marshal(token)
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(homeDir, tokenFilePath), jsonToken, 0600)
}

func loadOAuthToken() (*oauth2.Token, error) {
	homeDir, _ := os.UserHomeDir()
	jsonToken, err := os.ReadFile(filepath.Join(homeDir, tokenFilePath))
	if err != nil {
		return nil, err
	}

	var token oauth2.Token
	if err := json.Unmarshal(jsonToken, &token); err != nil {
		return nil, err
	}

	if token.Expiry.Before(time.Now()) {
		return &token, NewTokenExpiredError()
	}

	return &token, nil
}

func refreshToken(oldToken *oauth2.Token) *oauth2.Token {

	form := url.Values{}
	form.Add("grant_type", "refresh_token")
	form.Add("refresh_token", oldToken.RefreshToken)
	client_id := os.Getenv("SPOTIFY_ID")
	if client_id == "" {
		log.Fatalf("SPOTIFY_ID is not set")
	}
	form.Add("client_id", client_id)

	req, err := http.NewRequest("POST", "https://accounts.spotify.com/api/token", strings.NewReader(form.Encode()))
	if err != nil {
		log.Fatal(err)
		return nil
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
		return nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
		return nil
	}

	var newToken oauth2.Token
	if err := json.Unmarshal(body, &newToken); err != nil {
		log.Fatal(err)
		return nil
	}
	if newToken.Expiry.IsZero() {
		newToken.Expiry = time.Now().Add(time.Duration(3600) * time.Second)
	}

	return &newToken
}

func login() {
	verifier, _ := generateCodeVerifier()
	challenge := generateCodeChallenge(verifier)
	state := generateState()

	http.HandleFunc("/callback", completeAuth(state, verifier))
	go http.ListenAndServe(":21112", nil)

	url := auth.AuthURL(state,
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
		oauth2.SetAuthURLParam("code_challenge", challenge),
	)
	fmt.Println("Please log in to Spotify by visiting the following page in your browser:", url)

	browser.OpenURL(url)
}

func loginCmd() tea.Cmd {
	return func() tea.Msg {
		token, err := loadOAuthToken()

		if err != nil {
			if os.IsNotExist(err) {
				login()
				token = <-tokenCh
			} else if isTokenExpiredError(err) {
				token = refreshToken(token)
			} else {
				log.Fatal(err)
			}
			saveOAuthToken(token)
		}

		client := spotify.New(auth.Client(context.Background(), token))
		return AuthMsg{client}
	}
}

func completeAuth(state string, verifer string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tok, err := auth.Token(r.Context(), state, r,
			oauth2.SetAuthURLParam("code_verifier", verifer))
		if err != nil {
			http.Error(w, "Couldn't get token", http.StatusForbidden)
			log.Fatal(err)
		}
		if st := r.FormValue("state"); st != state {
			http.NotFound(w, r)
			log.Fatalf("State mismatch: %s != %s\n", st, state)
		}
		tokenCh <- tok
	}
}
