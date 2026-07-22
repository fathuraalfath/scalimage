package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"scalimage/internal/collage"
	"scalimage/internal/handler"
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

	// Initialize Collage Generator
	gen := collage.NewGenerator(store)

	// Initialize Handlers
	h := handler.NewHandler(store, gen)

	// Setup routing mux
	mux := http.NewServeMux()

	// API routes
	mux.HandleFunc("/api/upload", h.UploadHandler)
	mux.HandleFunc("/api/collage", h.CollageHandler)

	// Serve uploads directory
	mux.Handle("/uploads/", h.ServeUploadsHandler(absUploadDir))

	// CORS wrapper
	wrappedMux := handler.CorsMiddleware(mux)

	// Determine port
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting Scalimage backend on :%s", port)
	log.Printf("Serving uploaded files from %s", absUploadDir)

	if err := http.ListenAndServe(":"+port, wrappedMux); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
