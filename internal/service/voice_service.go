package service

import (
	"context"
	"errors"
	"time"

	"manjing-ai-go/internal/model"
	"manjing-ai-go/internal/repository"

	"gorm.io/gorm"
)

// VoiceService 声音服务接口
type VoiceService interface {
	Create(ctx context.Context, userID int64, req VoiceCreate) (*model.Voice, error)
	List(ctx context.Context, userID int64, query repository.VoiceListQuery) ([]model.Voice, int64, error)
	Get(ctx context.Context, userID, id int64) (*model.Voice, error)
	Update(ctx context.Context, userID, id int64, req VoiceUpdate) (*model.Voice, error)
	Delete(ctx context.Context, userID, id int64) error
}

// VoiceCreate 创建声音请求
type VoiceCreate struct {
	Name      string // 声音名称
	AgeGroup  int16  // 年龄段
	Gender    int16  // 性别
	Dialect   int16  // 方言口音
	Tone      int16  // 音色
	SampleURL string // 试听音频URL
	Type      int16  // 类型：1官方/2用户
}

// VoiceUpdate 更新声音请求
type VoiceUpdate struct {
	Name      *string // 声音名称
	AgeGroup  *int16  // 年龄段
	Gender    *int16  // 性别
	Dialect   *int16  // 方言口音
	Tone      *int16  // 音色
	SampleURL *string // 试听音频URL
}

// VoiceServiceImpl 声音服务实现
type VoiceServiceImpl struct {
	repo repository.VoiceRepository
}

// NewVoiceService 创建声音服务
func NewVoiceService(repo repository.VoiceRepository) *VoiceServiceImpl {
	return &VoiceServiceImpl{repo: repo}
}

func (s *VoiceServiceImpl) Create(ctx context.Context, userID int64, req VoiceCreate) (*model.Voice, error) {
	if userID == 0 {
		return nil, errors.New("未授权")
	}
	if req.Name == "" {
		return nil, errors.New("声音名称不能为空")
	}
	if req.AgeGroup < 1 || req.AgeGroup > 5 {
		return nil, errors.New("年龄段非法")
	}
	if req.Gender != 1 && req.Gender != 2 {
		return nil, errors.New("性别非法")
	}
	if req.Dialect < 1 || req.Dialect > 7 {
		return nil, errors.New("方言口音非法")
	}
	if req.Tone < 1 || req.Tone > 9 {
		return nil, errors.New("音色非法")
	}
	if req.Type != 1 && req.Type != 2 {
		return nil, errors.New("类型非法")
	}

	now := time.Now()
	voice := &model.Voice{
		Name:      req.Name,
		AgeGroup:  req.AgeGroup,
		Gender:    req.Gender,
		Dialect:   req.Dialect,
		Tone:      req.Tone,
		SampleURL: req.SampleURL,
		Type:      req.Type,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if req.Type == 2 {
		voice.OwnerUserID = &userID
	}

	if err := s.repo.Create(ctx, voice); err != nil {
		return nil, err
	}
	return voice, nil
}

func (s *VoiceServiceImpl) List(ctx context.Context, userID int64, query repository.VoiceListQuery) ([]model.Voice, int64, error) {
	if userID == 0 {
		return nil, 0, errors.New("未授权")
	}
	return s.repo.List(ctx, userID, query)
}

func (s *VoiceServiceImpl) Get(ctx context.Context, userID, id int64) (*model.Voice, error) {
	if userID == 0 {
		return nil, errors.New("未授权")
	}
	voice, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("声音不存在")
		}
		return nil, err
	}
	// 官方声音所有人可访问；用户声音仅创建者可访问
	if voice.Type == 2 && (voice.OwnerUserID == nil || *voice.OwnerUserID != userID) {
		return nil, errors.New("无权访问")
	}
	return voice, nil
}

func (s *VoiceServiceImpl) Update(ctx context.Context, userID, id int64, req VoiceUpdate) (*model.Voice, error) {
	if userID == 0 {
		return nil, errors.New("未授权")
	}
	voice, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("声音不存在")
		}
		return nil, err
	}
	// 官方声音仅管理员可更新（当前简化：所有人可更新官方声音 TODO: 后续增加管理员角色判断）
	if voice.Type == 2 && (voice.OwnerUserID == nil || *voice.OwnerUserID != userID) {
		return nil, errors.New("无权访问")
	}

	updates := map[string]interface{}{}
	if req.Name != nil {
		if *req.Name == "" {
			return nil, errors.New("声音名称不能为空")
		}
		updates["name"] = *req.Name
	}
	if req.AgeGroup != nil {
		if *req.AgeGroup < 1 || *req.AgeGroup > 5 {
			return nil, errors.New("年龄段非法")
		}
		updates["age_group"] = *req.AgeGroup
	}
	if req.Gender != nil {
		if *req.Gender != 1 && *req.Gender != 2 {
			return nil, errors.New("性别非法")
		}
		updates["gender"] = *req.Gender
	}
	if req.Dialect != nil {
		if *req.Dialect < 1 || *req.Dialect > 7 {
			return nil, errors.New("方言口音非法")
		}
		updates["dialect"] = *req.Dialect
	}
	if req.Tone != nil {
		if *req.Tone < 1 || *req.Tone > 9 {
			return nil, errors.New("音色非法")
		}
		updates["tone"] = *req.Tone
	}
	if req.SampleURL != nil {
		updates["sample_url"] = *req.SampleURL
	}
	if len(updates) == 0 {
		return voice, nil
	}

	updates["updated_at"] = time.Now()
	if err := s.repo.Update(ctx, id, updates); err != nil {
		return nil, err
	}
	return s.repo.FindByID(ctx, id)
}

func (s *VoiceServiceImpl) Delete(ctx context.Context, userID, id int64) error {
	if userID == 0 {
		return errors.New("未授权")
	}
	voice, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("声音不存在")
		}
		return err
	}
	if voice.Type == 2 && (voice.OwnerUserID == nil || *voice.OwnerUserID != userID) {
		return errors.New("无权访问")
	}

	now := time.Now()
	return s.repo.Update(ctx, id, map[string]interface{}{
		"deleted_at": &now,
		"updated_at": now,
	})
}
