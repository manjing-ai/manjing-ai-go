package service

import (
	"context"
	"errors"
	"time"

	"manjing-ai-go/internal/model"
	"manjing-ai-go/internal/repository"

	"gorm.io/gorm"
)

// ProjectService 项目服务
type ProjectService interface {
	Create(ctx context.Context, userID int64, name string, narrativeMode int16, coverResourceID *int64, videoAspectRatio, styleRef string) (*model.Project, error)
	List(ctx context.Context, userID int64, query repository.ProjectListQuery) ([]model.Project, int64, error)
	Get(ctx context.Context, userID, id int64) (*model.Project, error)
	Update(ctx context.Context, userID, id int64, req ProjectUpdate) (*model.Project, error)
	Delete(ctx context.Context, userID, id int64) error
	Restore(ctx context.Context, userID, id int64) (*model.Project, error)
	Archive(ctx context.Context, userID, id int64) (*model.Project, error)
}

// ProjectUpdate 更新请求
type ProjectUpdate struct {
	Name             *string
	NarrativeMode    *int16
	CoverResourceID  *int64
	VideoAspectRatio *string
	StyleRef         *string
	Status           *int16
}

// ProjectServiceImpl 实现
type ProjectServiceImpl struct {
	repo repository.ProjectRepository
}

// NewProjectService 创建服务
func NewProjectService(repo repository.ProjectRepository) *ProjectServiceImpl {
	return &ProjectServiceImpl{repo: repo}
}

func (s *ProjectServiceImpl) Create(ctx context.Context, userID int64, name string, narrativeMode int16, coverResourceID *int64, videoAspectRatio, styleRef string) (*model.Project, error) {
	if userID == 0 {
		return nil, errors.New("未授权")
	}
	if name == "" {
		return nil, errors.New("项目名称不能为空")
	}
	if narrativeMode == 0 {
		narrativeMode = 1
	}
	if narrativeMode != 1 && narrativeMode != 2 {
		return nil, errors.New("叙事模式非法")
	}
	if videoAspectRatio == "" {
		videoAspectRatio = "16:9"
	}

	now := time.Now()
	project := &model.Project{
		OwnerUserID:      userID,
		Name:             name,
		NarrativeMode:    narrativeMode,
		CoverResourceID:  coverResourceID,
		VideoAspectRatio: videoAspectRatio,
		StyleRef:         styleRef,
		Status:           1,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	if err := s.repo.Create(ctx, project); err != nil {
		return nil, err
	}
	return project, nil
}

func (s *ProjectServiceImpl) List(ctx context.Context, userID int64, query repository.ProjectListQuery) ([]model.Project, int64, error) {
	if userID == 0 {
		return nil, 0, errors.New("未授权")
	}
	return s.repo.List(ctx, userID, query)
}

func (s *ProjectServiceImpl) Get(ctx context.Context, userID, id int64) (*model.Project, error) {
	if userID == 0 {
		return nil, errors.New("未授权")
	}
	project, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("项目不存在")
		}
		return nil, err
	}
	if project.OwnerUserID != userID {
		return nil, errors.New("无权访问")
	}
	return project, nil
}

func (s *ProjectServiceImpl) Update(ctx context.Context, userID, id int64, req ProjectUpdate) (*model.Project, error) {
	if userID == 0 {
		return nil, errors.New("未授权")
	}
	project, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("项目不存在")
		}
		return nil, err
	}
	if project.OwnerUserID != userID {
		return nil, errors.New("无权访问")
	}

	updates := map[string]interface{}{}
	if req.Name != nil {
		if *req.Name == "" {
			return nil, errors.New("项目名称不能为空")
		}
		updates["name"] = *req.Name
	}
	if req.NarrativeMode != nil {
		if *req.NarrativeMode != 1 && *req.NarrativeMode != 2 {
			return nil, errors.New("叙事模式非法")
		}
		updates["narrative_mode"] = *req.NarrativeMode
	}
	if req.CoverResourceID != nil {
		updates["cover_resource_id"] = req.CoverResourceID
	}
	if req.VideoAspectRatio != nil {
		if *req.VideoAspectRatio == "" {
			return nil, errors.New("视频比例不能为空")
		}
		updates["video_aspect_ratio"] = *req.VideoAspectRatio
	}
	if req.StyleRef != nil {
		updates["style_ref"] = *req.StyleRef
	}
	if req.Status != nil {
		if *req.Status != 1 && *req.Status != 2 && *req.Status != 3 {
			return nil, errors.New("状态非法")
		}
		updates["status"] = *req.Status
		if *req.Status == 3 {
			now := time.Now()
			updates["deleted_at"] = &now
		} else {
			updates["deleted_at"] = nil
		}
	}
	if len(updates) == 0 {
		return project, nil
	}

	updates["updated_at"] = time.Now()
	if err := s.repo.Update(ctx, id, updates); err != nil {
		return nil, err
	}
	return s.repo.FindByID(ctx, id)
}

func (s *ProjectServiceImpl) Delete(ctx context.Context, userID, id int64) error {
	if userID == 0 {
		return errors.New("未授权")
	}
	project, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("项目不存在")
		}
		return err
	}
	if project.OwnerUserID != userID {
		return errors.New("无权访问")
	}
	if project.Status == 3 {
		return nil
	}
	now := time.Now()
	return s.repo.Update(ctx, id, map[string]interface{}{
		"status":     3,
		"deleted_at": &now,
		"updated_at": now,
	})
}

func (s *ProjectServiceImpl) Restore(ctx context.Context, userID, id int64) (*model.Project, error) {
	if userID == 0 {
		return nil, errors.New("未授权")
	}
	project, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("项目不存在")
		}
		return nil, err
	}
	if project.OwnerUserID != userID {
		return nil, errors.New("无权访问")
	}
	if project.Status != 3 {
		return project, nil
	}
	now := time.Now()
	if err := s.repo.Update(ctx, id, map[string]interface{}{
		"status":     1,
		"deleted_at": nil,
		"updated_at": now,
	}); err != nil {
		return nil, err
	}
	return s.repo.FindByID(ctx, id)
}

func (s *ProjectServiceImpl) Archive(ctx context.Context, userID, id int64) (*model.Project, error) {
	if userID == 0 {
		return nil, errors.New("未授权")
	}
	project, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("项目不存在")
		}
		return nil, err
	}
	if project.OwnerUserID != userID {
		return nil, errors.New("无权访问")
	}
	if project.Status == 2 {
		return project, nil
	}
	now := time.Now()
	if err := s.repo.Update(ctx, id, map[string]interface{}{
		"status":     2,
		"updated_at": now,
	}); err != nil {
		return nil, err
	}
	return s.repo.FindByID(ctx, id)
}
