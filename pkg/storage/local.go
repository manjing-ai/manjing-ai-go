package storage

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"manjing-ai-go/config"
)

// LocalStorage 本地存储实现
type LocalStorage struct {
	baseDir string
	baseURL string
}

// NewLocalStorage 创建本地存储
func NewLocalStorage(cfg config.LocalStorage) *LocalStorage {
	return &LocalStorage{
		baseDir: cfg.BaseDir,
		baseURL: cfg.BaseURL,
	}
}

func (s *LocalStorage) Save(ctx context.Context, objectKey string, data []byte) (*ObjectInfo, error) {
	path := filepath.Join(s.baseDir, filepath.FromSlash(objectKey))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return nil, err
	}
	return &ObjectInfo{ObjectKey: objectKey, URL: s.buildURL(objectKey)}, nil
}

func (s *LocalStorage) Delete(ctx context.Context, objectKey string) error {
	path := filepath.Join(s.baseDir, filepath.FromSlash(objectKey))
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func (s *LocalStorage) URL(ctx context.Context, objectKey string) (string, error) {
	return s.buildURL(objectKey), nil
}

func (s *LocalStorage) buildURL(objectKey string) string {
	base := strings.TrimRight(s.baseURL, "/")
	return base + "/" + strings.TrimLeft(objectKey, "/")
}
