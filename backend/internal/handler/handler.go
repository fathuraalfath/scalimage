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
	"scalimage/internal/compress"
	"scalimage/internal/resize"
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
	compress  *compress.Service
	resize    *resize.Service
}

// NewHandler creates a new Handler.
func NewHandler(store storage.Storage, generator *collage.Generator, comp *compress.Service, res *resize.Service) *Handler {
	return &Handler{
		store:     store,
		generator: generator,
		compress:  comp,
		resize:    res,
	}
}

// UploadHandler handles uploading multiple image files with strict security validations.
func (h *Handler) UploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Security: Limit total multipart body size to 32MB to prevent memory exhaustion
	r.Body = http.MaxBytesReader(w, r.Body, 32<<20)

	err := r.ParseMultipartForm(32 << 20)
	if err != nil {
		http.Error(w, "Failed to parse multipart form or file too large: "+err.Error(), http.StatusBadRequest)
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

		fileBytes, err := io.ReadAll(file)
		if err != nil {
			http.Error(w, "Failed to read file content: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Security: Strict MIME validation to prevent uploading malicious scripts or HTML
		mimeType := http.DetectContentType(fileBytes)
		if !strings.HasPrefix(mimeType, "image/") {
			http.Error(w, "Rejected non-image file type: "+mimeType, http.StatusBadRequest)
			return
		}

		// Parse dimensions & format
		config, format, err := image.DecodeConfig(bytes.NewReader(fileBytes))
		if err != nil {
			http.Error(w, "Invalid or unsupported image format: "+err.Error(), http.StatusBadRequest)
			return
		}

		// Security: Guard against image dimension bombs
		if config.Width <= 0 || config.Height <= 0 || config.Width > 8192 || config.Height > 8192 {
			http.Error(w, "Image dimensions exceed maximum supported resolution (8192x8192)", http.StatusBadRequest)
			return
		}

		// Generate random secure filename to avoid conflicts and directory traversal issues
		randBytes := make([]byte, 16)
		_, _ = rand.Read(randBytes)
		ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
		if ext == "" || ext == "." {
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

	// Security: Limit JSON request payload to 1MB
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var req collage.CollageRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Security: Guard canvas dimension boundaries
	if req.CanvasWidth <= 0 || req.CanvasHeight <= 0 || req.CanvasWidth > 8192 || req.CanvasHeight > 8192 {
		http.Error(w, "Canvas dimensions must be between 1 and 8192 px", http.StatusBadRequest)
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

	contentType := "image/png"
	if strings.ToLower(req.Format) == "jpeg" || strings.ToLower(req.Format) == "jpg" {
		contentType = "image/jpeg"
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Length", strconv.Itoa(buf.Len()))
	_, _ = buf.WriteTo(w)
}

// CompressHandler handles dedicated single-image compression and format conversion.
func (h *Handler) CompressHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var req compress.CompressRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	var buf bytes.Buffer
	if err := h.compress.Process(r.Context(), req, &buf); err != nil {
		http.Error(w, "Compression failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	contentType := "image/png"
	if strings.ToLower(req.Format) == "jpeg" || strings.ToLower(req.Format) == "jpg" {
		contentType = "image/jpeg"
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Length", strconv.Itoa(buf.Len()))
	_, _ = buf.WriteTo(w)
}

// ResizeHandler handles dedicated image dimension scaling.
func (h *Handler) ResizeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var req resize.ResizeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	var buf bytes.Buffer
	if err := h.resize.Process(r.Context(), req, &buf); err != nil {
		http.Error(w, "Resize failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

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

// CorsMiddleware injects CORS and HTTP security headers to protect against XSS and sniffing attacks.
func CorsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		// Security Hardening Headers
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
