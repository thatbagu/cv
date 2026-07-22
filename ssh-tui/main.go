package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	cssh "github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	bm "github.com/charmbracelet/wish/bubbletea"
	"github.com/muesli/termenv"
)

const (
	host    = ""
	port    = "2222"
	keyPath = ".ssh/cv_ssh_host_key"
)

// Loaded once at startup, shared across all sessions (read-only).
var (
	asciiFrames []string
	blogPosts   []BlogPost
)

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func loadFrames(dir string) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		log.Printf("frames not found at %s (set FRAMES_DIR to override)", dir)
		return nil
	}
	// Sort by filename
	var names []string
	for _, e := range entries {
		if !e.IsDir() {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)

	frames := make([]string, 0, len(names))
	for _, name := range names {
		data, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			continue
		}
		frames = append(frames, string(data))
	}
	log.Printf("loaded %d animation frames from %s", len(frames), dir)
	return frames
}

func teaHandler(s cssh.Session) (tea.Model, []tea.ProgramOption) {
	pty, _, active := s.Pty()
	if !active {
		fmt.Fprintln(s.Stderr(), "This service requires an interactive terminal.")
		s.Exit(1) //nolint
		return nil, nil
	}
	w, h := pty.Window.Width, pty.Window.Height
	if w == 0 {
		w = 80
	}
	if h == 0 {
		h = 24
	}
	renderer := bm.MakeRenderer(s)
	m := newModel(w, h, renderer, asciiFrames, blogPosts)
	return m, []tea.ProgramOption{tea.WithAltScreen()}
}

func main() {
	framesDir := envOr("FRAMES_DIR", "../assets/index/animation/cat")
	blogDir := envOr("BLOG_DIR", "../blog_posts")
	assetsDir := envOr("ASSETS_DIR", "../assets")

	asciiFrames = loadFrames(framesDir)
	blogPosts = loadBlogPosts(blogDir)
	loadImages(assetsDir)
	loadSteppeAssets(assetsDir + "/index/animation/steppe")
	log.Printf("loaded %d blog posts from %s", len(blogPosts), blogDir)
	log.Printf("loaded %d project images from %s", len(projectImgs), assetsDir+"/projects")

	s, err := wish.NewServer(
		wish.WithAddress(net.JoinHostPort(host, port)),
		wish.WithHostKeyPath(keyPath),
		wish.WithMiddleware(
			bm.MiddlewareWithColorProfile(teaHandler, termenv.TrueColor),
		),
	)
	if err != nil {
		log.Fatal("could not create server:", err)
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		if err := s.ListenAndServe(); err != nil {
			log.Println("server stopped:", err)
		}
	}()

	log.Printf("SSH TUI on port %s — connect: ssh -p %s <host>", port, port)

	sigch := make(chan os.Signal, 1)
	signal.Notify(sigch, os.Interrupt, syscall.SIGTERM)
	<-sigch

	log.Println("shutting down…")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := s.Shutdown(ctx); err != nil {
		log.Fatal("shutdown error:", err)
	}
	<-done
	log.Println("done")
}
