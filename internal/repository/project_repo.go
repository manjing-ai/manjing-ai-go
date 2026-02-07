package repository

import (
	"context"

	"manjing-ai-go/internal/model"

	"gorm.io/gorm"
)

// ProjectRepository 项目数据访问接口
type ProjectRepository interface {
	Create(ctx context.Context, project *model.Project) error
	Update(ctx context.Context, id int64, updates map[string]interface{}) error
	FindByID(ctx context.Context, id int64) (*model.Project, error)
	List(ctx context.Context, ownerUserID int64, query ProjectListQuery) ([]model.Project, int64, error)
}

// ProjectListQuery 项目列表查询
type ProjectListQuery struct {
	Page     int
	PageSize int
	Status   int
	Keyword  string
	Sort     string
}

// ProjectRepo 实现
type ProjectRepo struct {
	db *gorm.DB
}

// NewProjectRepo 创建仓库
func NewProjectRepo(db *gorm.DB) *ProjectRepo {
	return &ProjectRepo{db: db}
}

func (r *ProjectRepo) Create(ctx context.Context, project *model.Project) error {
	return r.db.WithContext(ctx).Create(project).Error
}

func (r *ProjectRepo) Update(ctx context.Context, id int64, updates map[string]interface{}) error {
	return r.db.WithContext(ctx).Model(&model.Project{}).Where("id = ?", id).Updates(updates).Error
}

func (r *ProjectRepo) FindByID(ctx context.Context, id int64) (*model.Project, error) {
	var project model.Project
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&project).Error; err != nil {
		return nil, err
	}
	return &project, nil
}

func (r *ProjectRepo) List(ctx context.Context, ownerUserID int64, query ProjectListQuery) ([]model.Project, int64, error) {
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
		query.Sort = "-updated_at"
	}

	db := r.db.WithContext(ctx).Model(&model.Project{}).Where("owner_user_id = ?", ownerUserID)
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
	case "created_at":
		db = db.Order("created_at ASC")
	case "-created_at":
		db = db.Order("created_at DESC")
	case "updated_at":
		db = db.Order("updated_at ASC")
	default:
		db = db.Order("updated_at DESC")
	}

	var items []model.Project
	err := db.Offset((query.Page - 1) * query.PageSize).Limit(query.PageSize).Find(&items).Error
	return items, total, err
}
