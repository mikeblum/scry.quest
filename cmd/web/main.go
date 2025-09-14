// Package main provides the web server for the Scry Quest D&D 5e assistant.
package main

import (
	"context"
	"embed"
	"flag"
	"io/fs"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/mikeblum/scry.quest/conf"
	"github.com/mikeblum/scry.quest/log"
	"github.com/mikeblum/scry.quest/templates"
)

//go:embed static/*
var staticFiles embed.FS

func main() {
	bind := flag.String("bind", "0.0.0.0:8080", "Address to bind the web server to (host:port or :port)")
	flag.Parse()

	config, err := conf.New(context.Background(), nil)
	if err != nil {
		slog.Error("Failed to load configuration", "error", err)
		return
	}
	log.NewFromEnv(config)

	mux := http.NewServeMux()

	// Static files
	staticFS, err := fs.Sub(staticFiles, "static")
	if err != nil {
		slog.Error("Failed to create static file system", "error", err)
		return
	}
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))

	// Routes
	mux.HandleFunc("/", handleHome)
	mux.HandleFunc("GET /api/chat/new", handleNewChat)
	mux.HandleFunc("POST /api/chat/new", handleNewChat)
	mux.HandleFunc("POST /api/chat/message", handleChatMessage)

	addr := *bind
	if strings.HasPrefix(addr, ":") {
		addr = "0.0.0.0" + addr
	}

	server := &http.Server{
		Addr:              addr,
		Handler:           corsMiddleware(loggingMiddleware(mux)),
		ReadHeaderTimeout: 5 * time.Second,
	}

	slog.Info("Starting server", "bind", addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("Server failed to start", "error", err)
	}
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(wrapped, r)

		duration := time.Since(start)
		slog.Info("HTTP request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", wrapped.statusCode,
			"duration", duration,
			"remote_addr", r.RemoteAddr,
			"user_agent", r.UserAgent(),
		)
	})
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func handleHome(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	component := templates.Layout("Scry Quest - D&D 5e Assistant")
	if err := component.Render(context.Background(), w); err != nil {
		slog.Error("Failed to render layout", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func handleNewChat(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	component := templates.ChatInterface()
	if err := component.Render(context.Background(), w); err != nil {
		slog.Error("Failed to render chat interface", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func handleChatMessage(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	message := r.FormValue("message")
	response := "I see you asked: " + message + ". This would integrate with your embeddings system to provide D&D 5e information."

	w.Header().Set("Content-Type", "text/html")

	// Return user message bubble followed by bot response
	userBubble := templates.MessageBubble(message, true)
	botBubble := templates.MessageBubble(response, false)

	if _, err := w.Write([]byte(`<div>`)); err != nil {
		slog.Error("Failed to write opening div", "error", err)
		return
	}
	if err := userBubble.Render(context.Background(), w); err != nil {
		slog.Error("Failed to render user bubble", "error", err)
		return
	}
	if err := botBubble.Render(context.Background(), w); err != nil {
		slog.Error("Failed to render bot bubble", "error", err)
		return
	}
	if _, err := w.Write([]byte(`</div>`)); err != nil {
		slog.Error("Failed to write closing div", "error", err)
		return
	}
}
