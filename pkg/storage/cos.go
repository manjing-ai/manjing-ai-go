package storage

import (
	"context"
	"errors"

	"manjing-ai-go/config"
)

// COSStorage 腾讯云 COS 占位实现
type COSStorage struct {
	bucket string
	region string
}

// NewCOSStorage 创建 COS 存储
func NewCOSStorage(cfg config.COSStorage) *COSStorage {
	return &COSStorage{
		bucket: cfg.Bucket,
		region: cfg.Region,
	}
}

func (s *COSStorage) Save(ctx context.Context, objectKey string, data []byte) (*ObjectInfo, error) {
	return nil, errors.New("cos storage not implemented")
}

func (s *COSStorage) Delete(ctx context.Context, objectKey string) error {
	return errors.New("cos storage not implemented")
}

func (s *COSStorage) URL(ctx context.Context, objectKey string) (string, error) {
	return "", errors.New("cos storage not implemented")
}
