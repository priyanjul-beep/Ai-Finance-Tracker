// Package storage provides a StorageService abstraction with a local-disk and
// Cloudinary implementation.  Select the provider via the STORAGE_PROVIDER env var.
package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// ─── Interface ────────────────────────────────────────────────────────────────

// Service abstracts file storage; both local and Cloudinary implement this.
type Service interface {
	Upload(ctx context.Context, filename string, data []byte, mimeType string) (string, error)
	Delete(ctx context.Context, fileURL string) error
	GetURL(ctx context.Context, path string) string
}

// ─── Local Storage ────────────────────────────────────────────────────────────

// LocalStorage stores files on the local filesystem.
type LocalStorage struct {
	basePath string
	baseURL  string
}

// NewLocalStorage creates a local storage provider.
// basePath: directory to store files (created if absent).
// baseURL: public URL prefix, e.g. "http://localhost:8080/uploads".
func NewLocalStorage(basePath, baseURL string) (*LocalStorage, error) {
	if err := os.MkdirAll(basePath, 0o755); err != nil {
		return nil, fmt.Errorf("local storage: mkdir: %w", err)
	}
	return &LocalStorage{basePath: basePath, baseURL: baseURL}, nil
}

// Upload writes data to disk under basePath/filename and returns the public URL.
func (l *LocalStorage) Upload(_ context.Context, filename string, data []byte, _ string) (string, error) {
	dest := filepath.Join(l.basePath, filename)
	if err := os.WriteFile(dest, data, 0o644); err != nil {
		return "", fmt.Errorf("local storage: write: %w", err)
	}
	return l.GetURL(nil, filename), nil
}

// Delete removes a file from disk.
func (l *LocalStorage) Delete(_ context.Context, fileURL string) error {
	// Derive filename from URL
	filename := filepath.Base(fileURL)
	dest := filepath.Join(l.basePath, filename)
	if err := os.Remove(dest); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("local storage: delete: %w", err)
	}
	return nil
}

// GetURL returns the public URL for a stored file.
func (l *LocalStorage) GetURL(_ context.Context, path string) string {
	return fmt.Sprintf("%s/%s", l.baseURL, filepath.Base(path))
}

// ─── Reader helper ────────────────────────────────────────────────────────────

// ReadAll reads all bytes from the given reader, useful for multipart uploads.
func ReadAll(r io.Reader) ([]byte, error) {
	return io.ReadAll(r)
}
