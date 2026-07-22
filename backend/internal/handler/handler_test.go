package handler

import (
	"bytes"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"scalimage/internal/collage"
	"scalimage/internal/storage"
)

func TestIntegration_UploadAndCollage(t *testing.T) {
	// Create temp dir for storage
	tempDir, err := os.MkdirTemp("", "scalimage-integration-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	store, err := storage.NewLocalStorage(tempDir)
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}

	gen := collage.NewGenerator(store)
	h := NewHandler(store, gen)

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/upload" {
			h.UploadHandler(w, r)
		} else if r.URL.Path == "/api/collage" {
			h.CollageHandler(w, r)
		} else {
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	// 1. Create a mock image file to upload
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			img.Set(x, y, color.RGBA{R: 255, G: 0, B: 0, A: 255})
		}
	}
	var imgBuf bytes.Buffer
	if err := png.Encode(&imgBuf, img); err != nil {
		t.Fatalf("failed to encode img: %v", err)
	}

	// 2. Perform upload request
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("images", "test.png")
	if err != nil {
		t.Fatalf("failed to create form file: %v", err)
	}
	if _, err := part.Write(imgBuf.Bytes()); err != nil {
		t.Fatalf("failed to write file to form part: %v", err)
	}
	writer.Close()

	req, err := http.NewRequest("POST", server.URL+"/api/upload", body)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("upload request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("upload returned non-OK status: %d", resp.StatusCode)
	}

	var uploadRes []UploadResponse
	if err := json.NewDecoder(resp.Body).Decode(&uploadRes); err != nil {
		t.Fatalf("failed to decode upload response: %v", err)
	}

	if len(uploadRes) != 1 {
		t.Fatalf("expected 1 uploaded image metadata, got %d", len(uploadRes))
	}
	uploaded := uploadRes[0]
	if uploaded.Width != 10 || uploaded.Height != 10 || uploaded.Format != "png" {
		t.Errorf("unexpected image metadata: %+v", uploaded)
	}

	// 3. Perform collage request
	collageReq := collage.CollageRequest{
		CanvasWidth:  100,
		CanvasHeight: 100,
		BGColor:      "#0000ff",
		Format:       "png",
		Images: []collage.PlacedImage{
			{
				ID:     uploaded.ID,
				X:      10,
				Y:      20,
				Width:  80,
				Height: 60,
			},
		},
	}
	reqBytes, err := json.Marshal(collageReq)
	if err != nil {
		t.Fatalf("failed to marshal collage request: %v", err)
	}

	req2, err := http.NewRequest("POST", server.URL+"/api/collage", bytes.NewReader(reqBytes))
	if err != nil {
		t.Fatalf("failed to create collage request: %v", err)
	}
	req2.Header.Set("Content-Type", "application/json")

	resp2, err := http.DefaultClient.Do(req2)
	if err != nil {
		t.Fatalf("collage request failed: %v", err)
	}
	defer resp2.Body.Close()

	if resp2.StatusCode != http.StatusOK {
		t.Fatalf("collage returned non-OK status: %d", resp2.StatusCode)
	}

	collageImg, err := png.Decode(resp2.Body)
	if err != nil {
		t.Fatalf("failed to decode returned collage image: %v", err)
	}

	bounds := collageImg.Bounds()
	if bounds.Dx() != 100 || bounds.Dy() != 100 {
		t.Errorf("expected 100x100 collage, got %dx%d", bounds.Dx(), bounds.Dy())
	}
}
