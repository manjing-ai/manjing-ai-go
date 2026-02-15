package repository

import (
	"context"

	"manjing-ai-go/internal/model"

	"gorm.io/gorm"
)

// VoiceRepository 声音数据访问接口
type VoiceRepository interface {
	Create(ctx context.Context, voice *model.Voice) error
	Update(ctx context.Context, id int64, updates map[string]interface{}) error
	FindByID(ctx context.Context, id int64) (*model.Voice, error)
	List(ctx context.Context, userID int64, query VoiceListQuery) ([]model.Voice, int64, error)
}

// VoiceListQuery 声音列表查询参数
type VoiceListQuery struct {
	Page     int
	PageSize int
	Type     int    // 类型筛选：1官方/2用户
	AgeGroup int    // 年龄段筛选
	Gender   int    // 性别筛选
	Dialect  int    // 方言口音筛选
	Tone     int    // 音色筛选
	Keyword  string // 名称关键词
	Sort     string // 排序
}

// VoiceRepo 声音仓库实现
type VoiceRepo struct {
	db *gorm.DB
}

// NewVoiceRepo 创建声音仓库
func NewVoiceRepo(db *gorm.DB) *VoiceRepo {
	return &VoiceRepo{db: db}
}

func (r *VoiceRepo) Create(ctx context.Context, voice *model.Voice) error {
	return r.db.WithContext(ctx).Create(voice).Error
}

func (r *VoiceRepo) Update(ctx context.Context, id int64, updates map[string]interface{}) error {
	return r.db.WithContext(ctx).Model(&model.Voice{}).Where("id = ?", id).Updates(updates).Error
}

func (r *VoiceRepo) FindByID(ctx context.Context, id int64) (*model.Voice, error) {
	var voice model.Voice
	if err := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&voice).Error; err != nil {
		return nil, err
	}
	return &voice, nil
}

func (r *VoiceRepo) List(ctx context.Context, userID int64, query VoiceListQuery) ([]model.Voice, int64, error) {
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 20
	}
	if query.PageSize > 100 {
		query.PageSize = 100
	}
	if query.Sort == "" {
		query.Sort = "-created_at"
	}

	db := r.db.WithContext(ctx).Model(&model.Voice{}).Where("deleted_at IS NULL")

	// 类型筛选：未指定则返回官方+当前用户的自定义声音
	switch query.Type {
	case 1:
		db = db.Where("type = ?", 1)
	case 2:
		db = db.Where("type = ? AND owner_user_id = ?", 2, userID)
	default:
		db = db.Where("(type = 1) OR (type = 2 AND owner_user_id = ?)", userID)
	}

	if query.AgeGroup > 0 {
		db = db.Where("age_group = ?", query.AgeGroup)
	}
	if query.Gender > 0 {
		db = db.Where("gender = ?", query.Gender)
	}
	if query.Dialect > 0 {
		db = db.Where("dialect = ?", query.Dialect)
	}
	if query.Tone > 0 {
		db = db.Where("tone = ?", query.Tone)
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

	var items []model.Voice
	err := db.Offset((query.Page - 1) * query.PageSize).Limit(query.PageSize).Find(&items).Error
	return items, total, err
}
