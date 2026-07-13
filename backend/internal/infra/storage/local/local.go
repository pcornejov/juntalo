// Package local implements app.FileStorage on local disk (Etapa 2 §2.1: R2 es
// la misma interfaz vía SDK S3-compatible, sin tocar el dominio).
package local

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

type Storage struct {
	rootDir   string
	publicURL string
}

func New(rootDir, publicURL string) *Storage {
	return &Storage{rootDir: rootDir, publicURL: publicURL}
}

func (s *Storage) Save(_ context.Context, key string, data []byte) error {
	path := filepath.Join(s.rootDir, filepath.FromSlash(key))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("local storage: mkdir: %w", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("local storage: write: %w", err)
	}
	return nil
}

func (s *Storage) PublicURL(key string) string {
	return fmt.Sprintf("%s/%s", s.publicURL, key)
}
