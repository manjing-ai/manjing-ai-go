package service

import (
	"context"
	"errors"
	"time"

	"manjing-ai-go/internal/model"
	"manjing-ai-go/internal/repository"

	"gorm.io/gorm"
)

// ChapterService 章节服务
type ChapterService interface {
	Create(ctx context.Context, userID int64, projectID int64, name, content, summary string, orderIndex int) (*model.Chapter, error)
	List(ctx context.Context, userID int64, projectID int64, query repository.ChapterListQuery) ([]model.Chapter, int64, error)
	Get(ctx context.Context, userID, id int64) (*model.Chapter, error)
	Update(ctx context.Context, userID, id int64, req ChapterUpdate) (*model.Chapter, error)
	Delete(ctx context.Context, userID, id int64) error
	Restore(ctx context.Context, userID, id int64) (*model.Chapter, error)
	Archive(ctx context.Context, userID, id int64) (*model.Chapter, error)
}

// ChapterUpdate 更新请求
type ChapterUpdate struct {
	Name       *string
	Content    *string
	Summary    *string
	OrderIndex *int
	Status     *int16
}

// ChapterServiceImpl 实现
type ChapterServiceImpl struct {
	repo        repository.ChapterRepository
	projectRepo repository.ProjectRepository
}

// NewChapterService 创建服务
func NewChapterService(repo repository.ChapterRepository, projectRepo repository.ProjectRepository) *ChapterServiceImpl {
	return &ChapterServiceImpl{repo: repo, projectRepo: projectRepo}
}

func (s *ChapterServiceImpl) Create(ctx context.Context, userID int64, projectID int64, name, content, summary string, orderIndex int) (*model.Chapter, error) {
	if userID == 0 {
		return nil, errors.New("未授权")
	}
	if projectID == 0 {
		return nil, errors.New("项目ID不能为空")
	}
	if name == "" {
		return nil, errors.New("章节名称不能为空")
	}
	project, err := s.projectRepo.FindByID(ctx, projectID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("项目不存在")
		}
		return nil, err
	}
	if project.OwnerUserID != userID {
		return nil, errors.New("无权访问")
	}

	now := time.Now()
	chapter := &model.Chapter{
		ProjectID:  projectID,
		Name:       name,
		Content:    content,
		Summary:    summary,
		OrderIndex: orderIndex,
		Status:     1,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if err := s.repo.Create(ctx, chapter); err != nil {
		return nil, err
	}
	return chapter, nil
}

func (s *ChapterServiceImpl) List(ctx context.Context, userID int64, projectID int64, query repository.ChapterListQuery) ([]model.Chapter, int64, error) {
	if userID == 0 {
		return nil, 0, errors.New("未授权")
	}
	if projectID == 0 {
		return nil, 0, errors.New("项目ID不能为空")
	}
	project, err := s.projectRepo.FindByID(ctx, projectID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, 0, errors.New("项目不存在")
		}
		return nil, 0, err
	}
	if project.OwnerUserID != userID {
		return nil, 0, errors.New("无权访问")
	}
	return s.repo.List(ctx, projectID, query)
}

func (s *ChapterServiceImpl) Get(ctx context.Context, userID, id int64) (*model.Chapter, error) {
	if userID == 0 {
		return nil, errors.New("未授权")
	}
	chapter, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("章节不存在")
		}
		return nil, err
	}
	project, err := s.projectRepo.FindByID(ctx, chapter.ProjectID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("项目不存在")
		}
		return nil, err
	}
	if project.OwnerUserID != userID {
		return nil, errors.New("无权访问")
	}
	return chapter, nil
}

func (s *ChapterServiceImpl) Update(ctx context.Context, userID, id int64, req ChapterUpdate) (*model.Chapter, error) {
	if userID == 0 {
		return nil, errors.New("未授权")
	}
	chapter, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("章节不存在")
		}
		return nil, err
	}
	project, err := s.projectRepo.FindByID(ctx, chapter.ProjectID)
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
			return nil, errors.New("章节名称不能为空")
		}
		updates["name"] = *req.Name
	}
	if req.Content != nil {
		updates["content"] = *req.Content
	}
	if req.Summary != nil {
		updates["summary"] = *req.Summary
	}
	if req.OrderIndex != nil {
		updates["order_index"] = *req.OrderIndex
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
		return chapter, nil
	}

	updates["updated_at"] = time.Now()
	if err := s.repo.Update(ctx, id, updates); err != nil {
		return nil, err
	}
	return s.repo.FindByID(ctx, id)
}

func (s *ChapterServiceImpl) Delete(ctx context.Context, userID, id int64) error {
	if userID == 0 {
		return errors.New("未授权")
	}
	chapter, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("章节不存在")
		}
		return err
	}
	project, err := s.projectRepo.FindByID(ctx, chapter.ProjectID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("项目不存在")
		}
		return err
	}
	if project.OwnerUserID != userID {
		return errors.New("无权访问")
	}
	if chapter.Status == 3 {
		return nil
	}
	now := time.Now()
	return s.repo.Update(ctx, id, map[string]interface{}{
		"status":     3,
		"deleted_at": &now,
		"updated_at": now,
	})
}

func (s *ChapterServiceImpl) Restore(ctx context.Context, userID, id int64) (*model.Chapter, error) {
	if userID == 0 {
		return nil, errors.New("未授权")
	}
	chapter, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("章节不存在")
		}
		return nil, err
	}
	project, err := s.projectRepo.FindByID(ctx, chapter.ProjectID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("项目不存在")
		}
		return nil, err
	}
	if project.OwnerUserID != userID {
		return nil, errors.New("无权访问")
	}
	if chapter.Status != 3 {
		return chapter, nil
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

func (s *ChapterServiceImpl) Archive(ctx context.Context, userID, id int64) (*model.Chapter, error) {
	if userID == 0 {
		return nil, errors.New("未授权")
	}
	chapter, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("章节不存在")
		}
		return nil, err
	}
	project, err := s.projectRepo.FindByID(ctx, chapter.ProjectID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("项目不存在")
		}
		return nil, err
	}
	if project.OwnerUserID != userID {
		return nil, errors.New("无权访问")
	}
	if chapter.Status == 2 {
		return chapter, nil
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
