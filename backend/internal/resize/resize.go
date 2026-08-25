package resize

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

	xdraw "golang.org/x/image/draw"
	_ "golang.org/x/image/webp"

	"scalimage/internal/storage"
)

// ResizeRequest defines target dimensions and format for an image.
type ResizeRequest struct {
	ID           string `json:"id"`
	TargetWidth  int    `json:"targetWidth"`
	TargetHeight int    `json:"targetHeight"`
	Format       string `json:"format"` // "png", "jpeg", "jpg"
}

// Service manages image resizing operations.
type Service struct {
	store storage.Storage
}

// NewService creates a new Resize Service.
func NewService(store storage.Storage) *Service {
	return &Service{store: store}
}

// Process reads an image from storage, resizes it using BiLinear interpolation, and writes it to out.
func (s *Service) Process(ctx context.Context, req ResizeRequest, out io.Writer) error {
	if req.ID == "" {
		return fmt.Errorf("image ID is required")
	}

	// Security & Resource Exhaustion guardrail: Bound target dimensions to [1, 8192]
	if req.TargetWidth <= 0 || req.TargetHeight <= 0 {
		return fmt.Errorf("target dimensions must be positive (got %dx%d)", req.TargetWidth, req.TargetHeight)
	}
	if req.TargetWidth > 8192 || req.TargetHeight > 8192 {
		return fmt.Errorf("target dimensions (%dx%d) exceed maximum supported limit of 8192x8192", req.TargetWidth, req.TargetHeight)
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

	// Guard against decompression bombs in source image
	srcBounds := srcImg.Bounds()
	if srcBounds.Dx() > 8192 || srcBounds.Dy() > 8192 {
		return fmt.Errorf("source image dimensions (%dx%d) exceed maximum limit of 8192x8192", srcBounds.Dx(), srcBounds.Dy())
	}

	dstRect := image.Rect(0, 0, req.TargetWidth, req.TargetHeight)
	dstImg := image.NewRGBA(dstRect)

	// High-quality BiLinear scaling
	xdraw.BiLinear.Scale(dstImg, dstRect, srcImg, srcBounds, xdraw.Over, nil)

	format := strings.ToLower(req.Format)
	if format == "jpeg" || format == "jpg" {
		return jpeg.Encode(out, dstImg, &jpeg.Options{Quality: 90})
	}

	return png.Encode(out, dstImg)
}
