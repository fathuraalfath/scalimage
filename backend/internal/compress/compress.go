package compress

import (
	"context"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"strings"

	_ "golang.org/x/image/webp"

	"scalimage/internal/storage"
)

// CompressRequest specifies compression and format conversion parameters.
type CompressRequest struct {
	ID      string `json:"id"`
	Format  string `json:"format"`  // "png", "jpeg", "jpg"
	Quality int    `json:"quality"` // 1-100 (applicable for JPEG)
}

// Service provides image compression functionality.
type Service struct {
	store storage.Storage
}

// NewService creates a new Compress Service.
func NewService(store storage.Storage) *Service {
	return &Service{store: store}
}

// Process reads an image from storage and re-encodes it with the requested format and quality.
func (s *Service) Process(ctx context.Context, req CompressRequest, out io.Writer) error {
	if req.ID == "" {
		return fmt.Errorf("image ID is required")
	}

	rc, err := s.store.Get(ctx, req.ID)
	if err != nil {
		return fmt.Errorf("failed to load image %q: %w", req.ID, err)
	}
	defer rc.Close()

	srcImg, _, err := image.Decode(rc)
	if err != nil {
		return fmt.Errorf("failed to decode image %q: %w", req.ID, err)
	}

	// Guard against decompression bombs / huge images
	bounds := srcImg.Bounds()
	if bounds.Dx() > 8192 || bounds.Dy() > 8192 {
		return fmt.Errorf("image dimensions (%dx%d) exceed maximum supported resolution of 8192x8192", bounds.Dx(), bounds.Dy())
	}

	format := strings.ToLower(req.Format)
	if format == "" {
		format = "jpeg"
	}

	if format == "jpeg" || format == "jpg" {
		q := req.Quality
		if q <= 0 || q > 100 {
			q = 80 // default quality
		}
		return jpeg.Encode(out, srcImg, &jpeg.Options{Quality: q})
	}

	// Default to PNG with standard compression
	return png.Encode(out, srcImg)
}
