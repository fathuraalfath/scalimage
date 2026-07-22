package storage

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLocalStorage_SaveAndGet(t *testing.T) {
	// Create a temporary directory for test uploads
	tempDir, err := os.MkdirTemp("", "scalimage-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	s, err := NewLocalStorage(tempDir)
	if err != nil {
		t.Fatalf("failed to init LocalStorage: %v", err)
	}

	ctx := context.Background()
	filename := "test_image.png"
	content := "fake-image-bytes-data"

	// 1. Test Save
	savedPath, err := s.Save(ctx, filename, strings.NewReader(content))
	if err != nil {
		t.Fatalf("failed to save file: %v", err)
	}

	expectedPath := filepath.Join(tempDir, filename)
	if savedPath != expectedPath {
		t.Errorf("expected saved path %q, got %q", expectedPath, savedPath)
	}

	// 2. Test Get
	rc, err := s.Get(ctx, filename)
	if err != nil {
		t.Fatalf("failed to get file: %v", err)
	}
	defer rc.Close()

	readData, err := io.ReadAll(rc)
	if err != nil {
		t.Fatalf("failed to read from reader: %v", err)
	}

	if string(readData) != content {
		t.Errorf("expected content %q, got %q", content, string(readData))
	}
}

func TestLocalStorage_DirectoryTraversalGuard(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "scalimage-test-traversal-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	s, err := NewLocalStorage(tempDir)
	if err != nil {
		t.Fatalf("failed to init LocalStorage: %v", err)
	}

	ctx := context.Background()

	// Try malicious paths
	maliciousKeys := []string{
		"../malicious.png",
		"sub/../../malicious.png",
		"/absolute/path/malicious.png",
	}

	for _, key := range maliciousKeys {
		_, err := s.Save(ctx, key, strings.NewReader("bad"))
		if err == nil {
			t.Errorf("expected traversal guard error for key %q, but got nil", key)
		}

		_, err = s.Get(ctx, key)
		if err == nil {
			t.Errorf("expected traversal guard error for key %q on Get, but got nil", key)
		}
	}
}
