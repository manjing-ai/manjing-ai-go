package repository

import (
	"context"

	"manjing-ai-go/internal/model"

	"gorm.io/gorm"
)

// LLMModelRepository LLM模型配置数据访问接口
type LLMModelRepository interface {
	Create(ctx context.Context, m *model.LLMModel) error
	Update(ctx context.Context, id int64, updates map[string]interface{}) error
	FindByID(ctx context.Context, id int64) (*model.LLMModel, error)
	FindActiveByPurpose(ctx context.Context, purpose string) (*model.LLMModel, error)
	List(ctx context.Context, query LLMModelListQuery) ([]model.LLMModel, int64, error)
}

// LLMModelListQuery 模型配置列表查询参数
type LLMModelListQuery struct {
	Page     int
	PageSize int
	Provider string
	Purpose  string
	IsActive *bool
}

// LLMModelRepo 模型配置仓库实现
type LLMModelRepo struct {
	db *gorm.DB
}

// NewLLMModelRepo 创建模型配置仓库
func NewLLMModelRepo(db *gorm.DB) *LLMModelRepo {
	return &LLMModelRepo{db: db}
}

func (r *LLMModelRepo) Create(ctx context.Context, m *model.LLMModel) error {
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *LLMModelRepo) Update(ctx context.Context, id int64, updates map[string]interface{}) error {
	return r.db.WithContext(ctx).Model(&model.LLMModel{}).Where("id = ?", id).Updates(updates).Error
}

func (r *LLMModelRepo) FindByID(ctx context.Context, id int64) (*model.LLMModel, error) {
	var m model.LLMModel
	if err := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&m).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

// FindActiveByPurpose 查找指定用途的启用模型
func (r *LLMModelRepo) FindActiveByPurpose(ctx context.Context, purpose string) (*model.LLMModel, error) {
	var m model.LLMModel
	err := r.db.WithContext(ctx).
		Where("purpose = ? AND is_active = true AND deleted_at IS NULL", purpose).
		Order("updated_at DESC").
		First(&m).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *LLMModelRepo) List(ctx context.Context, query LLMModelListQuery) ([]model.LLMModel, int64, error) {
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 20
	}
	if query.PageSize > 100 {
		query.PageSize = 100
	}

	db := r.db.WithContext(ctx).Model(&model.LLMModel{}).Where("deleted_at IS NULL")

	if query.Provider != "" {
		db = db.Where("provider = ?", query.Provider)
	}
	if query.Purpose != "" {
		db = db.Where("purpose = ?", query.Purpose)
	}
	if query.IsActive != nil {
		db = db.Where("is_active = ?", *query.IsActive)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var items []model.LLMModel
	err := db.Order("created_at DESC").
		Offset((query.Page - 1) * query.PageSize).
		Limit(query.PageSize).
		Find(&items).Error
	return items, total, err
}
