package storage

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
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

func TestLocalStorage_CleanupExpired(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "scalimage-test-cleanup-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	s, err := NewLocalStorage(tempDir)
	if err != nil {
		t.Fatalf("failed to init LocalStorage: %v", err)
	}

	ctx := context.Background()

	// 1. Create a dummy file
	oldFile := "old_image.png"
	path, err := s.Save(ctx, oldFile, strings.NewReader("old-bytes"))
	if err != nil {
		t.Fatalf("failed to save file: %v", err)
	}

	// 2. Backdate mod time to 48 hours ago
	oldTime := time.Now().Add(-48 * time.Hour)
	if err := os.Chtimes(path, oldTime, oldTime); err != nil {
		t.Fatalf("failed to backdate file: %v", err)
	}

	// 3. Create a recent file
	newFile := "new_image.png"
	if _, err := s.Save(ctx, newFile, strings.NewReader("new-bytes")); err != nil {
		t.Fatalf("failed to save new file: %v", err)
	}

	// 4. Run cleanup with 24h threshold
	deleted, err := s.CleanupExpired(24 * time.Hour)
	if err != nil {
		t.Fatalf("cleanup failed: %v", err)
	}
	if deleted != 1 {
		t.Errorf("expected 1 file deleted, got %d", deleted)
	}

	// 5. Verify old file is gone, new file remains
	if _, err := s.Get(ctx, oldFile); err == nil {
		t.Errorf("expected old file to be deleted, but still exists")
	}
	if rc, err := s.Get(ctx, newFile); err != nil {
		t.Errorf("expected new file to exist, got error: %v", err)
	} else {
		rc.Close()
	}
}
