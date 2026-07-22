package main

import (
	"encoding/json"
	"html"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

const (
	mastodonInstance = "https://mastodon.social"
	mastodonAccount  = "112950868045593874"
)

type MastodonPost struct {
	Text string
	Date string
	URL  string
}

type mastodonMsg struct {
	posts []MastodonPost
	err   error
}

var htmlTagRe = regexp.MustCompile(`<[^>]+>`)

func stripHTML(s string) string {
	s = htmlTagRe.ReplaceAllString(s, " ")
	s = html.UnescapeString(s)
	// Collapse whitespace
	fields := strings.Fields(s)
	return strings.Join(fields, " ")
}

func fetchMastodonCmd() tea.Cmd {
	return func() tea.Msg {
		url := mastodonInstance + "/api/v1/accounts/" + mastodonAccount +
			"/statuses?limit=3&exclude_replies=true&exclude_reblogs=true"

		client := &http.Client{Timeout: 8 * time.Second}
		resp, err := client.Get(url)
		if err != nil {
			return mastodonMsg{err: err}
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return mastodonMsg{err: err}
		}

		var raw []struct {
			Content   string `json:"content"`
			CreatedAt string `json:"created_at"`
			URL       string `json:"url"`
		}
		if err := json.Unmarshal(body, &raw); err != nil {
			return mastodonMsg{err: err}
		}

		posts := make([]MastodonPost, 0, len(raw))
		for _, r := range raw {
			t, err := time.Parse(time.RFC3339, r.CreatedAt)
			date := r.CreatedAt
			if err == nil {
				date = t.Format("2006-01-02")
			}
			posts = append(posts, MastodonPost{
				Text: stripHTML(r.Content),
				Date: date,
				URL:  r.URL,
			})
		}
		return mastodonMsg{posts: posts}
	}
}
