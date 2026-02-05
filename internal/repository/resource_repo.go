package repository

import (
	"context"
	"errors"
	"time"

	"manjing-ai-go/internal/model"

	"gorm.io/gorm"
)

// ResourceRepository 资源数据访问接口
type ResourceRepository interface {
	Create(ctx context.Context, res *model.Resource) error
	Update(ctx context.Context, id int64, updates map[string]interface{}) error
	FindByID(ctx context.Context, id int64) (*model.Resource, error)
	List(ctx context.Context, userID int64, query ResourceListQuery) ([]model.Resource, int64, error)
	SoftDelete(ctx context.Context, id int64) error
	HardDelete(ctx context.Context, id int64) error
	SumSizeByUser(ctx context.Context, userID int64) (int64, error)
}

// ResourceListQuery 列表查询
type ResourceListQuery struct {
	Page     int
	PageSize int
	Type     string
	Category string
	Status   string
	Keyword  string
	Sort     string
}

// ResourceRepo 实现
type ResourceRepo struct {
	db *gorm.DB
}

// NewResourceRepo 创建仓库
func NewResourceRepo(db *gorm.DB) *ResourceRepo {
	return &ResourceRepo{db: db}
}

func (r *ResourceRepo) Create(ctx context.Context, res *model.Resource) error {
	return r.db.WithContext(ctx).Create(res).Error
}

func (r *ResourceRepo) Update(ctx context.Context, id int64, updates map[string]interface{}) error {
	return r.db.WithContext(ctx).Model(&model.Resource{}).Where("id = ?", id).Updates(updates).Error
}

func (r *ResourceRepo) FindByID(ctx context.Context, id int64) (*model.Resource, error) {
	var res model.Resource
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&res).Error
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func (r *ResourceRepo) List(ctx context.Context, userID int64, query ResourceListQuery) ([]model.Resource, int64, error) {
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 20
	}
	if query.PageSize > 100 {
		query.PageSize = 100
	}
	if query.Status == "" {
		query.Status = "active"
	}
	if query.Sort == "" {
		query.Sort = "-created_at"
	}

	db := r.db.WithContext(ctx).Model(&model.Resource{}).Where("user_id = ?", userID)
	if query.Type != "" {
		db = db.Where("type = ?", query.Type)
	}
	if query.Category != "" {
		db = db.Where("category = ?", query.Category)
	}
	if query.Status != "" {
		db = db.Where("status = ?", query.Status)
	}
	if query.Keyword != "" {
		db = db.Where("name ILIKE ?", "%"+query.Keyword+"%")
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	switch query.Sort {
	case "created_at":
		db = db.Order("created_at ASC")
	default:
		db = db.Order("created_at DESC")
	}

	var items []model.Resource
	err := db.Offset((query.Page - 1) * query.PageSize).Limit(query.PageSize).Find(&items).Error
	return items, total, err
}

func (r *ResourceRepo) SoftDelete(ctx context.Context, id int64) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&model.Resource{}).Where("id = ?", id).
		Updates(map[string]interface{}{"status": "deleted", "deleted_at": now}).Error
}

func (r *ResourceRepo) HardDelete(ctx context.Context, id int64) error {
	res := r.db.WithContext(ctx).Delete(&model.Resource{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("not found")
	}
	return nil
}

func (r *ResourceRepo) SumSizeByUser(ctx context.Context, userID int64) (int64, error) {
	var sum int64
	err := r.db.WithContext(ctx).Model(&model.Resource{}).Where("user_id = ?", userID).Select("COALESCE(SUM(size_bytes),0)").Scan(&sum).Error
	return sum, err
}
