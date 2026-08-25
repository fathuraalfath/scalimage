package compress

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
	"testing"

	"scalimage/internal/storage"
)

func TestCompressService_Process(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "scalimage-compress-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	store, err := storage.NewLocalStorage(tempDir)
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}

	ctx := context.Background()

	// 1. Create a dummy PNG
	img := image.NewRGBA(image.Rect(0, 0, 50, 50))
	for y := 0; y < 50; y++ {
		for x := 0; x < 50; x++ {
			img.Set(x, y, color.RGBA{R: 200, G: 100, B: 50, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("failed to encode test img: %v", err)
	}

	testID := "test_sample.png"
	if _, err := store.Save(ctx, testID, &buf); err != nil {
		t.Fatalf("failed to save test img: %v", err)
	}

	svc := NewService(store)

	// 2. Test compress to JPEG with 50% quality
	var jpegOut bytes.Buffer
	err = svc.Process(ctx, CompressRequest{
		ID:      testID,
		Format:  "jpeg",
		Quality: 50,
	}, &jpegOut)
	if err != nil {
		t.Fatalf("compress to jpeg failed: %v", err)
	}

	decodedJpeg, err := jpeg.Decode(&jpegOut)
	if err != nil {
		t.Fatalf("failed to decode output jpeg: %v", err)
	}
	if decodedJpeg.Bounds().Dx() != 50 || decodedJpeg.Bounds().Dy() != 50 {
		t.Errorf("expected 50x50 dimensions, got %dx%d", decodedJpeg.Bounds().Dx(), decodedJpeg.Bounds().Dy())
	}

	// 3. Test compress to PNG
	var pngOut bytes.Buffer
	err = svc.Process(ctx, CompressRequest{
		ID:     testID,
		Format: "png",
	}, &pngOut)
	if err != nil {
		t.Fatalf("compress to png failed: %v", err)
	}

	decodedPng, err := png.Decode(&pngOut)
	if err != nil {
		t.Fatalf("failed to decode output png: %v", err)
	}
	if decodedPng.Bounds().Dx() != 50 || decodedPng.Bounds().Dy() != 50 {
		t.Errorf("expected 50x50 dimensions, got %dx%d", decodedPng.Bounds().Dx(), decodedPng.Bounds().Dy())
	}
}
