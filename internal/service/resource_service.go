package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"manjing-ai-go/config"
	"manjing-ai-go/internal/model"
	"manjing-ai-go/internal/repository"
	"manjing-ai-go/pkg/storage"

	"gorm.io/datatypes"
)

// ResourceService 资源服务
type ResourceService interface {
	Upload(ctx context.Context, userID int64, fileName string, fileBytes []byte, name, resType, category, extraData string) (*model.Resource, string, error)
	List(ctx context.Context, userID int64, query repository.ResourceListQuery) ([]ResourceListItem, int64, error)
	Get(ctx context.Context, userID, id int64) (*model.Resource, string, error)
	Update(ctx context.Context, userID, id int64, name, category, extraData string) (*model.Resource, error)
	Delete(ctx context.Context, userID, id int64, hard bool) error
}

// ResourceServiceImpl 实现
type ResourceServiceImpl struct {
	repo    repository.ResourceRepository
	storage storage.Service
	cfg     config.StorageConfig
}

// ResourceListItem 列表项
type ResourceListItem struct {
	Resource model.Resource
	URL      string
}

// NewResourceService 创建服务
func NewResourceService(repo repository.ResourceRepository, storageSvc storage.Service, cfg config.StorageConfig) *ResourceServiceImpl {
	return &ResourceServiceImpl{repo: repo, storage: storageSvc, cfg: cfg}
}

func (s *ResourceServiceImpl) Upload(ctx context.Context, userID int64, fileName string, fileBytes []byte, name, resType, category, extraData string) (*model.Resource, string, error) {
	if userID == 0 {
		return nil, "", errors.New("未授权")
	}
	if len(fileBytes) == 0 {
		return nil, "", errors.New("文件为空")
	}

	maxFile := s.cfg.MaxFileSizeMB * 1024 * 1024
	if maxFile > 0 && int64(len(fileBytes)) > maxFile {
		return nil, "", errors.New("文件过大")
	}
	total, err := s.repo.SumSizeByUser(ctx, userID)
	if err != nil {
		return nil, "", err
	}
	maxTotal := s.cfg.MaxTotalSizeMB * 1024 * 1024
	if maxTotal > 0 && total+int64(len(fileBytes)) > maxTotal {
		return nil, "", errors.New("超过总配额限制")
	}

	mimeType := http.DetectContentType(fileBytes[:min(512, len(fileBytes))])
	if resType == "" {
		resType = detectType(mimeType)
	}

	if name == "" {
		name = strings.TrimSuffix(fileName, filepath.Ext(fileName))
	}

	ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(fileName)), ".")
	width, height := 0, 0
	if strings.HasPrefix(mimeType, "image/") {
		if cfg, _, err := image.DecodeConfig(bytes.NewReader(fileBytes)); err == nil {
			width, height = cfg.Width, cfg.Height
		}
	}

	aspect := ""
	if width > 0 && height > 0 {
		aspect = fmt.Sprintf("%d:%d", width, height)
	}

	extra := datatypes.JSON([]byte("null"))
	if extraData != "" {
		var tmp interface{}
		if err := json.Unmarshal([]byte(extraData), &tmp); err != nil {
			return nil, "", errors.New("extra_data 格式错误")
		}
		extra = datatypes.JSON([]byte(extraData))
	}

	res := &model.Resource{
		UserID:    userID,
		Name:      name,
		Type:      resType,
		Category:  category,
		ObjectKey: "pending",
		FileName:  fileName,
		FileExt:   ext,
		MimeType:  mimeType,
		Width:     width,
		Height:    height,
		Aspect:    aspect,
		SizeBytes: int64(len(fileBytes)),
		Status:    "pending",
		ExtraData: extra,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := s.repo.Create(ctx, res); err != nil {
		return nil, "", err
	}

	objectKey := buildObjectKey(userID, resType, res.ID, ext)
	info, err := s.storage.Save(ctx, objectKey, fileBytes)
	if err != nil {
		_ = s.repo.HardDelete(ctx, res.ID)
		return nil, "", err
	}

	updates := map[string]interface{}{
		"object_key": objectKey,
		"status":     "active",
		"updated_at": time.Now(),
	}
	if err := s.repo.Update(ctx, res.ID, updates); err != nil {
		return nil, "", err
	}

	res.ObjectKey = objectKey
	res.Status = "active"
	res.UpdatedAt = time.Now()
	return res, info.URL, nil
}

func (s *ResourceServiceImpl) List(ctx context.Context, userID int64, query repository.ResourceListQuery) ([]ResourceListItem, int64, error) {
	items, total, err := s.repo.List(ctx, userID, query)
	if err != nil {
		return nil, 0, err
	}
	result := make([]ResourceListItem, 0, len(items))
	for _, it := range items {
		url, err := s.storage.URL(ctx, it.ObjectKey)
		if err != nil {
			return nil, 0, err
		}
		result = append(result, ResourceListItem{Resource: it, URL: url})
	}
	return result, total, nil
}

func (s *ResourceServiceImpl) Get(ctx context.Context, userID, id int64) (*model.Resource, string, error) {
	res, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, "", err
	}
	if res.UserID != userID {
		return nil, "", errors.New("无权访问")
	}
	url, err := s.storage.URL(ctx, res.ObjectKey)
	if err != nil {
		return nil, "", err
	}
	return res, url, nil
}

func (s *ResourceServiceImpl) Update(ctx context.Context, userID, id int64, name, category, extraData string) (*model.Resource, error) {
	res, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if res.UserID != userID {
		return nil, errors.New("无权访问")
	}

	updates := map[string]interface{}{}
	if name != "" {
		updates["name"] = name
	}
	if category != "" {
		updates["category"] = category
	}
	if extraData != "" {
		var tmp interface{}
		if err := json.Unmarshal([]byte(extraData), &tmp); err != nil {
			return nil, errors.New("extra_data 格式错误")
		}
		updates["extra_data"] = datatypes.JSON([]byte(extraData))
	}
	if len(updates) == 0 {
		return res, nil
	}
	updates["updated_at"] = time.Now()
	if err := s.repo.Update(ctx, id, updates); err != nil {
		return nil, err
	}
	return s.repo.FindByID(ctx, id)
}

func (s *ResourceServiceImpl) Delete(ctx context.Context, userID, id int64, hard bool) error {
	res, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if res.UserID != userID {
		return errors.New("无权访问")
	}

	if hard {
		if err := s.storage.Delete(ctx, res.ObjectKey); err != nil {
			return err
		}
		return s.repo.HardDelete(ctx, id)
	}
	return s.repo.SoftDelete(ctx, id)
}

func buildObjectKey(userID int64, resType string, id int64, ext string) string {
	ts := time.Now().Format("20060102_150405")
	if ext != "" {
		return fmt.Sprintf("user_%d/%s/%d_%s.%s", userID, resType, id, ts, ext)
	}
	return fmt.Sprintf("user_%d/%s/%d_%s", userID, resType, id, ts)
}

func detectType(mimeType string) string {
	if strings.HasPrefix(mimeType, "image/") {
		return "image"
	}
	if strings.HasPrefix(mimeType, "video/") {
		return "video"
	}
	if strings.HasPrefix(mimeType, "audio/") {
		return "audio"
	}
	return "other"
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
