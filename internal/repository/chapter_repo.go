package repository

import (
	"context"

	"manjing-ai-go/internal/model"

	"gorm.io/gorm"
)

// ChapterRepository 章节数据访问接口
type ChapterRepository interface {
	Create(ctx context.Context, chapter *model.Chapter) error
	Update(ctx context.Context, id int64, updates map[string]interface{}) error
	FindByID(ctx context.Context, id int64) (*model.Chapter, error)
	List(ctx context.Context, projectID int64, query ChapterListQuery) ([]model.Chapter, int64, error)
}

// ChapterListQuery 章节列表查询
type ChapterListQuery struct {
	Page     int
	PageSize int
	Status   int
	Keyword  string
	Sort     string
}

// ChapterRepo 实现
type ChapterRepo struct {
	db *gorm.DB
}

// NewChapterRepo 创建仓库
func NewChapterRepo(db *gorm.DB) *ChapterRepo {
	return &ChapterRepo{db: db}
}

func (r *ChapterRepo) Create(ctx context.Context, chapter *model.Chapter) error {
	return r.db.WithContext(ctx).Create(chapter).Error
}

func (r *ChapterRepo) Update(ctx context.Context, id int64, updates map[string]interface{}) error {
	return r.db.WithContext(ctx).Model(&model.Chapter{}).Where("id = ?", id).Updates(updates).Error
}

func (r *ChapterRepo) FindByID(ctx context.Context, id int64) (*model.Chapter, error) {
	var chapter model.Chapter
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&chapter).Error; err != nil {
		return nil, err
	}
	return &chapter, nil
}

func (r *ChapterRepo) List(ctx context.Context, projectID int64, query ChapterListQuery) ([]model.Chapter, int64, error) {
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
		query.Sort = "order_index"
	}

	db := r.db.WithContext(ctx).Model(&model.Chapter{}).Where("project_id = ?", projectID)
	if query.Status > 0 {
		db = db.Where("status = ?", query.Status)
	} else {
		db = db.Where("status <> ?", 3)
	}
	if query.Keyword != "" {
		db = db.Where("name ILIKE ?", "%"+query.Keyword+"%")
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	switch query.Sort {
	case "-order_index":
		db = db.Order("order_index DESC")
	case "created_at":
		db = db.Order("created_at ASC")
	case "-created_at":
		db = db.Order("created_at DESC")
	default:
		db = db.Order("order_index ASC")
	}

	var items []model.Chapter
	err := db.Offset((query.Page - 1) * query.PageSize).Limit(query.PageSize).Find(&items).Error
	return items, total, err
}
