package storage

import "context"

// ObjectInfo 存储对象信息
type ObjectInfo struct {
	ObjectKey string
	URL       string
}

// Service 存储服务接口
type Service interface {
	Save(ctx context.Context, objectKey string, data []byte) (*ObjectInfo, error)
	Delete(ctx context.Context, objectKey string) error
	URL(ctx context.Context, objectKey string) (string, error)
}
