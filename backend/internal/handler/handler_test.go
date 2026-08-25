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
	"scalimage/internal/compress"
	"scalimage/internal/resize"
	"scalimage/internal/storage"
)

func TestIntegration_UploadCollageCompressResize(t *testing.T) {
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
	comp := compress.NewService(store)
	res := resize.NewService(store)
	h := NewHandler(store, gen, comp, res)

	// Create test server with CorsMiddleware
	mux := http.NewServeMux()
	mux.HandleFunc("/api/upload", h.UploadHandler)
	mux.HandleFunc("/api/collage", h.CollageHandler)
	mux.HandleFunc("/api/compress", h.CompressHandler)
	mux.HandleFunc("/api/resize", h.ResizeHandler)

	server := httptest.NewServer(CorsMiddleware(mux))
	defer server.Close()

	// 1. Create a mock image file to upload
	img := image.NewRGBA(image.Rect(0, 0, 20, 20))
	for y := 0; y < 20; y++ {
		for x := 0; x < 20; x++ {
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

	// Verify security headers
	if resp.Header.Get("X-Content-Type-Options") != "nosniff" {
		t.Errorf("expected X-Content-Type-Options: nosniff header")
	}

	var uploadRes []UploadResponse
	if err := json.NewDecoder(resp.Body).Decode(&uploadRes); err != nil {
		t.Fatalf("failed to decode upload response: %v", err)
	}

	if len(uploadRes) != 1 {
		t.Fatalf("expected 1 uploaded image metadata, got %d", len(uploadRes))
	}
	uploaded := uploadRes[0]

	// 3. Test /api/compress endpoint
	compressBody, _ := json.Marshal(compress.CompressRequest{
		ID:      uploaded.ID,
		Format:  "jpeg",
		Quality: 75,
	})
	reqCompress, _ := http.NewRequest("POST", server.URL+"/api/compress", bytes.NewReader(compressBody))
	reqCompress.Header.Set("Content-Type", "application/json")
	respCompress, err := http.DefaultClient.Do(reqCompress)
	if err != nil || respCompress.StatusCode != http.StatusOK {
		t.Fatalf("compress request failed (status: %d)", respCompress.StatusCode)
	}
	respCompress.Body.Close()

	// 4. Test /api/resize endpoint
	resizeBody, _ := json.Marshal(resize.ResizeRequest{
		ID:           uploaded.ID,
		TargetWidth:  40,
		TargetHeight: 40,
		Format:       "png",
	})
	reqResize, _ := http.NewRequest("POST", server.URL+"/api/resize", bytes.NewReader(resizeBody))
	reqResize.Header.Set("Content-Type", "application/json")
	respResize, err := http.DefaultClient.Do(reqResize)
	if err != nil || respResize.StatusCode != http.StatusOK {
		t.Fatalf("resize request failed (status: %d)", respResize.StatusCode)
	}
	respResize.Body.Close()
}
