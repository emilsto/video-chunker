package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/emilsto/video-chunker/internal/api"
	"github.com/emilsto/video-chunker/internal/config"
)

func main() {
	fmt.Println("Starting Video Chunker API Server")

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	port := cfg.Server.Port
	if port == "" {
		port = "5000"
	}
	server := &http.Server{
		Addr:    ":" + port,
		Handler: api.SetupRoutes(),
	}

	go func() {
		fmt.Printf("Server listening on port %s\n", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	fmt.Println("Shutting down server...")

	// TODO: graceful shutdown
	fmt.Println("Server stopped")
}
