package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"scalimage/internal/collage"
	"scalimage/internal/compress"
	"scalimage/internal/handler"
	"scalimage/internal/resize"
	"scalimage/internal/storage"
)

func main() {
	// Base upload directory
	uploadDir := "./uploads"

	// Ensure uploadDir is absolute
	absUploadDir, err := filepath.Abs(uploadDir)
	if err != nil {
		log.Fatalf("Failed to resolve absolute upload path: %v", err)
	}

	// Initialize Storage
	store, err := storage.NewLocalStorage(absUploadDir)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}

	// Start background storage cleanup (purges uploads older than 24h every 1h)
	// ponytail: Background goroutine tied to root context; avoids external cron dependency.
	cleanupCtx, cleanupCancel := context.WithCancel(context.Background())
	defer cleanupCancel()
	store.StartCleanupRoutine(cleanupCtx, 24*time.Hour, 1*time.Hour)

	// Initialize Services
	gen := collage.NewGenerator(store)
	comp := compress.NewService(store)
	res := resize.NewService(store)

	// Initialize Handlers
	h := handler.NewHandler(store, gen, comp, res)

	// Setup routing mux
	mux := http.NewServeMux()

	// API routes
	mux.HandleFunc("/api/upload", h.UploadHandler)
	mux.HandleFunc("/api/collage", h.CollageHandler)
	mux.HandleFunc("/api/compress", h.CompressHandler)
	mux.HandleFunc("/api/resize", h.ResizeHandler)

	// Serve uploads directory
	mux.Handle("/uploads/", h.ServeUploadsHandler(absUploadDir))

	// CORS & Security wrapper
	wrappedMux := handler.CorsMiddleware(mux)

	// Determine port
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Configure hardened production HTTP server with explicit timeouts
	server := &http.Server{
		Addr:              ":" + port,
		Handler:           wrappedMux,
		ReadHeaderTimeout: 5 * time.Second,  // Protects against Slowloris attacks
		ReadTimeout:       30 * time.Second, // Max time to read entire request including upload stream
		WriteTimeout:      60 * time.Second, // Max time to write response stream
		IdleTimeout:       120 * time.Second,
	}

	// Run server in background goroutine to enable graceful shutdown
	go func() {
		log.Printf("Starting Scalimage backend on :%s", port)
		log.Printf("Serving uploaded files from %s", absUploadDir)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Graceful shutdown listener
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)
	<-stopChan

	log.Println("Shutting down server gracefully...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}
	log.Println("Server stopped cleanly.")
}
