package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"guessthedish/internal/content"
	"guessthedish/internal/game"
	"guessthedish/internal/httpapi"
)

func main() {
	contentPath := os.Getenv("CONTENT_PATH")
	if contentPath == "" {
		contentPath = content.DefaultPath
	}
	bundle, err := content.Load(contentPath)
	if err != nil {
		log.Fatal(err)
	}

	addr := os.Getenv("ADDR")
	if addr == "" {
		addr = "127.0.0.1:8080"
	}
	distPath := os.Getenv("DIST_PATH")
	if distPath == "" {
		distPath = "dist"
	}
	server := &http.Server{
		Addr:              addr,
		Handler:           httpapi.New(game.NewStore(bundle.Puzzles), content.Catalog(bundle), distPath),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	shutdown := make(chan os.Signal, 1)
	shutdownComplete := make(chan struct{})
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		defer close(shutdownComplete)
		<-shutdown
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			log.Printf("graceful shutdown: %v", err)
		}
	}()

	log.Printf("serving on http://%s with %d puzzles", addr, len(bundle.Puzzles))
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
	<-shutdownComplete
}
