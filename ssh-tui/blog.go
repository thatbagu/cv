package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/charmbracelet/glamour"
	"gopkg.in/yaml.v3"
)

type BlogPost struct {
	Title   string
	Date    string
	Excerpt string
	Slug    string
	Body    string
}

type frontmatter struct {
	Title   string `yaml:"title"`
	Date    string `yaml:"date"`
	Excerpt string `yaml:"excerpt"`
	Slug    string `yaml:"slug"`
}

func loadBlogPosts(dir string) []BlogPost {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var posts []BlogPost
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		post, err := parseBlogPost(filepath.Join(dir, e.Name()))
		if err != nil {
			continue
		}
		posts = append(posts, post)
	}
	sort.Slice(posts, func(i, j int) bool {
		return posts[i].Date > posts[j].Date
	})
	return posts
}

func parseBlogPost(path string) (BlogPost, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return BlogPost{}, err
	}
	parts := strings.SplitN(string(data), "---", 3)
	if len(parts) < 3 {
		return BlogPost{}, fmt.Errorf("no frontmatter in %s", path)
	}
	var fm frontmatter
	if err := yaml.Unmarshal([]byte(parts[1]), &fm); err != nil {
		return BlogPost{}, err
	}
	return BlogPost{
		Title:   fm.Title,
		Date:    fm.Date,
		Excerpt: fm.Excerpt,
		Slug:    fm.Slug,
		Body:    strings.TrimSpace(parts[2]),
	}, nil
}

func renderBlogList(posts []BlogPost, cursor, w int, st styles) string {
	if len(posts) == 0 {
		return st.sub.Render("\n  No blog posts found.\n")
	}
	var sb strings.Builder
	sb.WriteString("\n")
	for i, p := range posts {
		if i == cursor {
			sb.WriteString(st.peach.Bold(true).Render("  ❯ "+p.Title) + "\n")
			sb.WriteString(st.sub.Render("    "+p.Date) + "\n")
			if p.Excerpt != "" {
				sb.WriteString(st.body.Width(w-8).PaddingLeft(4).Render(p.Excerpt) + "\n")
			}
		} else {
			sb.WriteString(st.body.Render("    "+p.Title) + "\n")
			sb.WriteString(st.sub.Render("    "+p.Date) + "\n")
		}
		sb.WriteString("\n")
	}
	sb.WriteString(st.sub.Render("  enter: read post") + "\n")
	return sb.String()
}

func renderBlogPost(post BlogPost, w int, st styles) string {
	var sb strings.Builder
	sb.WriteString(fmtH1(post.Title, st))
	sb.WriteString(st.sub.Render("  "+post.Date) + "\n\n")

	r, err := glamour.NewTermRenderer(
		glamour.WithStylePath("dark"),
		glamour.WithWordWrap(w-4),
	)
	if err != nil {
		sb.WriteString(st.body.Width(w-4).PaddingLeft(2).Render(post.Body))
		return sb.String()
	}
	rendered, err := r.Render(post.Body)
	if err != nil {
		sb.WriteString(st.body.Width(w-4).PaddingLeft(2).Render(post.Body))
		return sb.String()
	}
	sb.WriteString(rendered)
	return sb.String()
}
