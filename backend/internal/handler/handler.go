package handler

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	_ "golang.org/x/image/webp"

	"scalimage/internal/collage"
	"scalimage/internal/storage"
)

// UploadResponse represents the metadata of an uploaded file.
type UploadResponse struct {
	ID     string `json:"id"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
	Format string `json:"format"`
}

// Handler contains dependencies for HTTP routes.
type Handler struct {
	store     storage.Storage
	generator *collage.Generator
}

// NewHandler creates a new Handler.
func NewHandler(store storage.Storage, generator *collage.Generator) *Handler {
	return &Handler{
		store:     store,
		generator: generator,
	}
}

// UploadHandler handles uploading multiple image files.
func (h *Handler) UploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Limit to 32MB uploads
	err := r.ParseMultipartForm(32 << 20)
	if err != nil {
		http.Error(w, "Failed to parse multipart form: "+err.Error(), http.StatusBadRequest)
		return
	}

	files := r.MultipartForm.File["images"]
	if len(files) == 0 {
		http.Error(w, "No files found under key 'images'", http.StatusBadRequest)
		return
	}

	var responses []UploadResponse

	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			http.Error(w, "Failed to open file: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer file.Close()

		// Read bytes into memory to parse dimensions and save
		fileBytes, err := io.ReadAll(file)
		if err != nil {
			http.Error(w, "Failed to read file content: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Parse dimensions
		config, format, err := image.DecodeConfig(bytes.NewReader(fileBytes))
		if err != nil {
			http.Error(w, "Invalid image file format: "+err.Error(), http.StatusBadRequest)
			return
		}

		// Generate random secure filename to avoid conflicts and directory traversal issues
		randBytes := make([]byte, 16)
		_, _ = rand.Read(randBytes)
		ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
		if ext == "" {
			// Fallback extension based on decoded format
			ext = "." + format
		}
		uniqueID := hex.EncodeToString(randBytes) + ext

		// Save to storage
		_, err = h.store.Save(r.Context(), uniqueID, bytes.NewReader(fileBytes))
		if err != nil {
			http.Error(w, "Failed to save file to storage: "+err.Error(), http.StatusInternalServerError)
			return
		}

		responses = append(responses, UploadResponse{
			ID:     uniqueID,
			Width:  config.Width,
			Height: config.Height,
			Format: format,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(responses)
}

// CollageHandler parses layout requests, executes the composition engine, and writes the binary image stream.
func (h *Handler) CollageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req collage.CollageRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Default values
	if req.BGColor == "" {
		req.BGColor = "#ffffff"
	}
	if req.Format == "" {
		req.Format = "png"
	}

	var buf bytes.Buffer
	err = h.generator.Generate(r.Context(), req, &buf)
	if err != nil {
		http.Error(w, "Collage generation failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Return binary stream
	contentType := "image/png"
	if strings.ToLower(req.Format) == "jpeg" || strings.ToLower(req.Format) == "jpg" {
		contentType = "image/jpeg"
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Length", strconv.Itoa(buf.Len()))
	_, _ = buf.WriteTo(w)
}

// ServeUploadsHandler returns a handler that serves the raw uploaded assets.
func (h *Handler) ServeUploadsHandler(baseDir string) http.Handler {
	return http.StripPrefix("/uploads/", http.FileServer(http.Dir(baseDir)))
}

// CorsMiddleware injects CORS headers to facilitate local multi-port environment requests.
func CorsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
