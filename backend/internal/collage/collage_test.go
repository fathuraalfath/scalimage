package collage

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

func TestParseHexColor(t *testing.T) {
	tests := []struct {
		input    string
		expected color.RGBA
		wantErr  bool
	}{
		{"#ffffff", color.RGBA{R: 255, G: 255, B: 255, A: 255}, false},
		{"#000", color.RGBA{R: 0, G: 0, B: 0, A: 255}, false},
		{"ff0000", color.RGBA{R: 255, G: 0, B: 0, A: 255}, false},
		{"#invalid", color.RGBA{}, true},
		{"#12", color.RGBA{}, true},
	}

	for _, tc := range tests {
		res, err := ParseHexColor(tc.input)
		if (err != nil) != tc.wantErr {
			t.Errorf("ParseHexColor(%q) error status unexpected; got err=%v, wantErr=%v", tc.input, err, tc.wantErr)
			continue
		}
		if !tc.wantErr && res != tc.expected {
			t.Errorf("ParseHexColor(%q) = %+v, expected %+v", tc.input, res, tc.expected)
		}
	}
}

func TestGenerator_Generate(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "scalimage-collage-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	s, err := storage.NewLocalStorage(tempDir)
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}

	ctx := context.Background()

	// 1. Create and save a simple 10x10 red PNG image
	redImg := image.NewRGBA(image.Rect(0, 0, 10, 10))
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			redImg.Set(x, y, color.RGBA{R: 255, G: 0, B: 0, A: 255})
		}
	}
	var redBuf bytes.Buffer
	if err := png.Encode(&redBuf, redImg); err != nil {
		t.Fatalf("failed to encode red img: %v", err)
	}
	redID := "red.png"
	if _, err := s.Save(ctx, redID, &redBuf); err != nil {
		t.Fatalf("failed to save red img: %v", err)
	}

	// 2. Create and save a simple 10x10 blue PNG image
	blueImg := image.NewRGBA(image.Rect(0, 0, 10, 10))
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			blueImg.Set(x, y, color.RGBA{R: 0, G: 0, B: 255, A: 255})
		}
	}
	var blueBuf bytes.Buffer
	if err := png.Encode(&blueBuf, blueImg); err != nil {
		t.Fatalf("failed to encode blue img: %v", err)
	}
	blueID := "blue.png"
	if _, err := s.Save(ctx, blueID, &blueBuf); err != nil {
		t.Fatalf("failed to save blue img: %v", err)
	}

	// 3. Initialize Generator and request a 100x100 collage containing both images
	gen := NewGenerator(s)
	req := CollageRequest{
		CanvasWidth:  100,
		CanvasHeight: 100,
		BGColor:      "#00ff00", // green background
		Format:       "png",
		Images: []PlacedImage{
			{ID: redID, X: 0, Y: 0, Width: 50, Height: 100},
			{ID: blueID, X: 50, Y: 0, Width: 50, Height: 100},
		},
	}

	var outputBuf bytes.Buffer
	if err := gen.Generate(ctx, req, &outputBuf); err != nil {
		t.Fatalf("collage generation failed: %v", err)
	}

	// 4. Verify generated image dimensions
	collageImg, err := png.Decode(&outputBuf)
	if err != nil {
		t.Fatalf("failed to decode generated collage: %v", err)
	}

	bounds := collageImg.Bounds()
	if bounds.Dx() != 100 || bounds.Dy() != 100 {
		t.Errorf("expected output dimensions 100x100, got %dx%d", bounds.Dx(), bounds.Dy())
	}
}
