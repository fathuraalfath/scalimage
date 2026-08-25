package resize

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/png"
	"os"
	"testing"

	"scalimage/internal/storage"
)

func TestResizeService_Process(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "scalimage-resize-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	store, err := storage.NewLocalStorage(tempDir)
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}

	ctx := context.Background()

	// 1. Create a 100x100 source image
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			img.Set(x, y, color.RGBA{R: 50, G: 150, B: 250, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("failed to encode test img: %v", err)
	}

	testID := "resize_sample.png"
	if _, err := store.Save(ctx, testID, &buf); err != nil {
		t.Fatalf("failed to save test img: %v", err)
	}

	svc := NewService(store)

	// 2. Test valid downscale to 40x30
	var out bytes.Buffer
	err = svc.Process(ctx, ResizeRequest{
		ID:           testID,
		TargetWidth:  40,
		TargetHeight: 30,
		Format:       "png",
	}, &out)
	if err != nil {
		t.Fatalf("resize process failed: %v", err)
	}

	decoded, err := png.Decode(&out)
	if err != nil {
		t.Fatalf("failed to decode resized img: %v", err)
	}
	if decoded.Bounds().Dx() != 40 || decoded.Bounds().Dy() != 30 {
		t.Errorf("expected 40x30 dimensions, got %dx%d", decoded.Bounds().Dx(), decoded.Bounds().Dy())
	}

	// 3. Test invalid negative / zero dimensions
	err = svc.Process(ctx, ResizeRequest{
		ID:           testID,
		TargetWidth:  -10,
		TargetHeight: 50,
	}, &bytes.Buffer{})
	if err == nil {
		t.Errorf("expected error for negative width, got nil")
	}

	// 4. Test oversized dimension guardrail (>8192)
	err = svc.Process(ctx, ResizeRequest{
		ID:           testID,
		TargetWidth:  10000,
		TargetHeight: 10000,
	}, &bytes.Buffer{})
	if err == nil {
		t.Errorf("expected error for oversized dimension (>8192), got nil")
	}
}
