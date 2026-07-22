package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var sectionNames = [5]string{"Home", "About", "Projects", "Blog", "Contact"}

type tickMsg time.Time

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second/24, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

type styles struct {
	h1, h2, h3  lipgloss.Style
	body, sub   lipgloss.Style
	green, sap  lipgloss.Style
	lav, peach  lipgloss.Style
	activeTab   lipgloss.Style
	inactiveTab lipgloss.Style
	sep         lipgloss.Style
	footer      lipgloss.Style
}

func newStyles(r *lipgloss.Renderer) styles {
	return styles{
		h1:          r.NewStyle().Foreground(lipgloss.Color("#fab387")).Bold(true),
		h2:          r.NewStyle().Foreground(lipgloss.Color("#cba6f7")).Bold(true),
		h3:          r.NewStyle().Foreground(lipgloss.Color("#89b4fa")).Bold(true),
		body:        r.NewStyle().Foreground(lipgloss.Color("#cdd6f4")),
		sub:         r.NewStyle().Foreground(lipgloss.Color("#a6adc8")),
		green:       r.NewStyle().Foreground(lipgloss.Color("#a6e3a1")),
		sap:         r.NewStyle().Foreground(lipgloss.Color("#74c7ec")),
		lav:         r.NewStyle().Foreground(lipgloss.Color("#b4befe")),
		peach:       r.NewStyle().Foreground(lipgloss.Color("#fab387")),
		activeTab:   r.NewStyle().Background(lipgloss.Color("#313244")).Foreground(lipgloss.Color("#fab387")).Bold(true),
		inactiveTab: r.NewStyle().Background(lipgloss.Color("#181825")).Foreground(lipgloss.Color("#a6adc8")),
		sep:         r.NewStyle().Foreground(lipgloss.Color("#fab387")),
		footer:      r.NewStyle().Background(lipgloss.Color("#313244")).Foreground(lipgloss.Color("#a6adc8")),
	}
}

type blogView int

const (
	blogList blogView = iota
	blogPost
)

type model struct {
	active      int
	vp          viewport.Model
	width       int
	height      int
	st          styles
	frames      []string
	frameIdx    int
	steppe      *steppeScene
	blogPosts   []BlogPost
	blogCursor      int
	blogView        blogView
	mastodonPosts   []MastodonPost
	mastodonLoading bool
	searching       bool
	searchInput     textinput.Model
	searchResults   []searchResult
	searchCursor    int
}

func newModel(w, h int, renderer *lipgloss.Renderer, frames []string, posts []BlogPost) model {
	st := newStyles(renderer)

	vpH := h - 3
	vp := viewport.New(w, vpH)
	vp.KeyMap.HalfPageDown = key.NewBinding(key.WithKeys("d", "ctrl+d"))
	vp.KeyMap.HalfPageUp = key.NewBinding(key.WithKeys("u", "ctrl+u"))

	ti := textinput.New()
	ti.Placeholder = "search posts and projects…"
	ti.CharLimit = 80
	ti.Width = max(20, w-12)

	m := model{
		active:          0,
		vp:              vp,
		width:           w,
		height:          h,
		st:              st,
		frames:          frames,
		steppe:          newSteppeScene(w),
		blogPosts:       posts,
		mastodonLoading: true,
		searchInput:     ti,
	}
	m.vp.SetContent(m.sectionContent())
	return m
}

func (m model) Init() tea.Cmd {
	cmds := []tea.Cmd{fetchMastodonCmd()}
	if len(m.frames) > 0 {
		cmds = append(cmds, tickCmd())
	}
	return tea.Batch(cmds...)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.vp.Width = msg.Width
		m.vp.Height = msg.Height - 3
		m.searchInput.Width = max(20, msg.Width-12)
		m.vp.SetContent(m.sectionContent())
		return m, nil

	case tickMsg:
		tickSteppeScene(m.steppe)
		if len(m.frames) > 0 {
			m.frameIdx = (m.frameIdx + 1) % len(m.frames)
		}
		if m.active == 0 {
			m.vp.SetContent(m.sectionContent())
		}
		return m, tickCmd()

	case mastodonMsg:
		m.mastodonLoading = false
		m.mastodonPosts = msg.posts
		if m.active == 0 {
			m.vp.SetContent(m.sectionContent())
		}
		return m, nil

	case tea.KeyMsg:
		// Search mode (home only)
		if m.active == 0 && m.searching {
			switch msg.String() {
			case "esc":
				m.searching = false
				m.searchInput.Reset()
				m.searchResults = nil
				m.vp.SetContent(m.sectionContent())
				return m, nil
			case "enter":
				if len(m.searchResults) > 0 {
					return m.jumpToResult(m.searchResults[m.searchCursor]), nil
				}
				return m, nil
			case "up":
				if m.searchCursor > 0 {
					m.searchCursor--
					m.vp.SetContent(m.sectionContent())
				}
				return m, nil
			case "down":
				if m.searchCursor < len(m.searchResults)-1 {
					m.searchCursor++
					m.vp.SetContent(m.sectionContent())
				}
				return m, nil
			default:
				var cmd tea.Cmd
				m.searchInput, cmd = m.searchInput.Update(msg)
				m.searchResults = doSearch(m.searchInput.Value(), m.blogPosts)
				m.searchCursor = 0
				m.vp.SetContent(m.sectionContent())
				return m, cmd
			}
		}
		// Activate search from home with /
		if m.active == 0 && msg.String() == "/" {
			m.searching = true
			cmd := m.searchInput.Focus()
			m.vp.SetContent(m.sectionContent())
			return m, cmd
		}

		// Blog post view
		if m.active == 3 && m.blogView == blogPost {
			switch msg.String() {
			case "esc", "backspace":
				m.blogView = blogList
				m.vp.GotoTop()
				m.vp.SetContent(m.sectionContent())
				return m, nil
			case "q", "Q", "ctrl+c":
				return m, tea.Quit
			default:
				var cmd tea.Cmd
				m.vp, cmd = m.vp.Update(msg)
				return m, cmd
			}
		}

		// Blog list view
		if m.active == 3 && m.blogView == blogList {
			switch msg.String() {
			case "q", "Q", "ctrl+c":
				return m, tea.Quit
			case "j", "down":
				if m.blogCursor < len(m.blogPosts)-1 {
					m.blogCursor++
					m.vp.SetContent(m.sectionContent())
				}
				return m, nil
			case "k", "up":
				if m.blogCursor > 0 {
					m.blogCursor--
					m.vp.SetContent(m.sectionContent())
				}
				return m, nil
			case "enter":
				if len(m.blogPosts) > 0 {
					m.blogView = blogPost
					m.vp.GotoTop()
					m.vp.SetContent(m.sectionContent())
				}
				return m, nil
			}
			// fall through to section navigation below
		}

		// All sections: shared navigation
		switch msg.String() {
		case "q", "Q", "ctrl+c":
			return m, tea.Quit
		case "1", "2", "3", "4", "5":
			return m.goTo(int(msg.String()[0]-'1')), nil
		case "right", "tab", "l":
			return m.goTo((m.active + 1) % len(sectionNames)), nil
		case "left", "shift+tab", "h":
			return m.goTo((m.active - 1 + len(sectionNames)) % len(sectionNames)), nil
		case "g":
			m.vp.GotoTop()
			return m, nil
		case "G":
			m.vp.GotoBottom()
			return m, nil
		default:
			var cmd tea.Cmd
			m.vp, cmd = m.vp.Update(msg)
			return m, cmd
		}
	}
	return m, nil
}

func (m model) goTo(idx int) model {
	m.active = idx
	m.searching = false
	m.searchInput.Blur()
	m.blogView = blogList
	m.vp.GotoTop()
	m.vp.SetContent(m.sectionContent())
	return m
}

func (m model) jumpToResult(r searchResult) model {
	m.searching = false
	m.searchInput.Reset()
	m.searchInput.Blur()
	m.searchResults = nil
	m.active = r.section
	m.blogView = blogList
	if r.section == 3 && r.blogIdx >= 0 {
		m.blogCursor = r.blogIdx
		m.blogView = blogPost
	}
	m.vp.GotoTop()
	m.vp.SetContent(m.sectionContent())
	return m
}

func (m model) sectionContent() string {
	switch m.active {
	case 0:
		return renderHome(m.width, m.st, m.frames, m.frameIdx, m.steppe,
			m.mastodonPosts, m.mastodonLoading,
			m.searching, m.searchInput.View(), m.searchInput.Value(),
			m.searchResults, m.searchCursor)
	case 1:
		return renderAbout(m.width, m.st)
	case 2:
		return renderProjects(m.width, m.st)
	case 3:
		if m.blogView == blogPost {
			return renderBlogPost(m.blogPosts[m.blogCursor], m.width, m.st)
		}
		return renderBlogList(m.blogPosts, m.blogCursor, m.width, m.st)
	case 4:
		return renderContact(m.width, m.st)
	}
	return ""
}

func (m model) View() string {
	if m.width == 0 {
		return "\n  Loading…"
	}
	return m.header() + "\n" + m.separator() + "\n" + m.vp.View() + "\n" + m.footer()
}

func (m model) header() string {
	var tabW int
	var tabs []string
	for i, name := range sectionNames {
		label := fmt.Sprintf(" %d:%s ", i+1, name)
		tabW += len(label)
		if i == m.active {
			tabs = append(tabs, m.st.activeTab.Render(label))
		} else {
			tabs = append(tabs, m.st.inactiveTab.Render(label))
		}
	}
	pad := m.st.inactiveTab.Render(strings.Repeat(" ", max(0, m.width-tabW)))
	return strings.Join(tabs, "") + pad
}

func (m model) separator() string {
	return m.st.sep.Render(strings.Repeat("─", m.width))
}

func (m model) footer() string {
	var keys string
	switch {
	case m.active == 0 && m.searching:
		keys = " type to search  ↑↓: select  enter: go  esc: cancel"
	case m.active == 3 && m.blogView == blogPost:
		keys = " ↑↓/jk: scroll  esc: back to list  q: quit"
	case m.active == 3:
		keys = " j/k: select post  enter: read  tab/←→: section  q: quit"
	default:
		keys = " ↑↓/jk: scroll  d/u: ½page  /: search  tab/←→: nav  q: quit"
	}
	pct := fmt.Sprintf("%.0f%%", m.vp.ScrollPercent()*100)
	pad := strings.Repeat(" ", max(0, m.width-len(keys)-len(pct)-1))
	return m.st.footer.Render(keys + pad + pct + " ")
}
