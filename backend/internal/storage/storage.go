package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Storage defines interface for managing file uploads.
type Storage interface {
	Save(ctx context.Context, key string, data io.Reader) (string, error)
	Get(ctx context.Context, key string) (io.ReadCloser, error)
	GetFullPath(key string) (string, error)
	CleanupExpired(maxAge time.Duration) (int, error)
}

// LocalStorage implements Storage on local filesystem.
type LocalStorage struct {
	baseDir string
}

// NewLocalStorage creates and initializes a LocalStorage.
func NewLocalStorage(baseDir string) (*LocalStorage, error) {
	absDir, err := filepath.Abs(baseDir)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path of base directory: %w", err)
	}

	if err := os.MkdirAll(absDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create base directory: %w", err)
	}

	return &LocalStorage{baseDir: absDir}, nil
}

// sanitizeKey checks for directory traversal and returns absolute path.
func (s *LocalStorage) sanitizeKey(key string) (string, error) {
	// Clean the path to resolve any `..`
	cleanedKey := filepath.Clean(key)
	if strings.HasPrefix(cleanedKey, "..") || filepath.IsAbs(cleanedKey) || strings.HasPrefix(cleanedKey, "/") || strings.HasPrefix(cleanedKey, "\\") {
		return "", errors.New("invalid key: directory traversal attempt")
	}

	// Join with baseDir and verify it still resides within baseDir
	targetPath := filepath.Join(s.baseDir, cleanedKey)
	absPath, err := filepath.Abs(targetPath)
	if err != nil {
		return "", fmt.Errorf("invalid path resolution: %w", err)
	}

	if !strings.HasPrefix(absPath, s.baseDir) {
		return "", errors.New("invalid key: resolved path is outside base directory")
	}

	return absPath, nil
}

// Save stores the data with key and returns the absolute file path.
func (s *LocalStorage) Save(ctx context.Context, key string, data io.Reader) (string, error) {
	targetPath, err := s.sanitizeKey(key)
	if err != nil {
		return "", err
	}

	// Ensure subdirectories inside baseDir are created
	if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
		return "", fmt.Errorf("failed to create directories for key: %w", err)
	}

	file, err := os.Create(targetPath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	if _, err := io.Copy(file, data); err != nil {
		return "", fmt.Errorf("failed to save file content: %w", err)
	}

	return targetPath, nil
}

// Get retrieves the file as a ReadCloser.
func (s *LocalStorage) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	targetPath, err := s.sanitizeKey(key)
	if err != nil {
		return nil, err
	}

	file, err := os.Open(targetPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	return file, nil
}

// GetFullPath returns the absolute path of the file.
func (s *LocalStorage) GetFullPath(key string) (string, error) {
	return s.sanitizeKey(key)
}

// CleanupExpired deletes files older than maxAge in the base directory, skipping .gitkeep.
// ponytail: Simple O(N) scan using standard filepath.Walk; sufficient for local directories.
func (s *LocalStorage) CleanupExpired(maxAge time.Duration) (int, error) {
	if maxAge <= 0 {
		return 0, nil
	}

	cutoff := time.Now().Add(-maxAge)
	deleted := 0

	err := filepath.Walk(s.baseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || info.Name() == ".gitkeep" {
			return nil
		}
		if info.ModTime().Before(cutoff) {
			if removeErr := os.Remove(path); removeErr == nil {
				deleted++
			}
		}
		return nil
	})

	return deleted, err
}

// StartCleanupRoutine periodically sweeps and removes expired upload files in the background.
func (s *LocalStorage) StartCleanupRoutine(ctx context.Context, maxAge, interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				count, err := s.CleanupExpired(maxAge)
				if err != nil {
					log.Printf("[Storage Cleanup] Error during sweep: %v", err)
				} else if count > 0 {
					log.Printf("[Storage Cleanup] Purged %d expired files", count)
				}
			}
		}
	}()
}
