package service

import (
	"context"
	"errors"
	"time"

	"manjing-ai-go/config"
	"manjing-ai-go/internal/model"
	"manjing-ai-go/internal/repository"
	"manjing-ai-go/pkg/llm"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// LLMService LLM服务接口
type LLMService interface {
	// 模型配置 CRUD
	CreateModel(ctx context.Context, req LLMModelCreate) (*model.LLMModel, error)
	ListModels(ctx context.Context, query repository.LLMModelListQuery) ([]model.LLMModel, int64, error)
	GetModel(ctx context.Context, id int64) (*model.LLMModel, error)
	UpdateModel(ctx context.Context, id int64, req LLMModelUpdate) (*model.LLMModel, error)
	DeleteModel(ctx context.Context, id int64) error
	// 对话
	Chat(ctx context.Context, userID int64, req LLMChatRequest) (*LLMChatResponse, error)
	// 日志
	ListLogs(ctx context.Context, query repository.LLMCallLogListQuery) ([]model.LLMCallLog, int64, error)
	LogStats(ctx context.Context, query repository.LLMCallLogStatsQuery) (*repository.LLMCallLogStats, []repository.LLMCallLogGroupStats, error)
}

// LLMModelCreate 创建模型配置请求
type LLMModelCreate struct {
	Name        string  `json:"name"`
	Provider    string  `json:"provider"`
	BaseURL     string  `json:"base_url"`
	APIKey      string  `json:"api_key"`
	Model       string  `json:"model"`
	MaxTokens   int     `json:"max_tokens"`
	Temperature float64 `json:"temperature"`
	Timeout     int     `json:"timeout"`
	Purpose     string  `json:"purpose"`
}

// LLMModelUpdate 更新模型配置请求
type LLMModelUpdate struct {
	Name        *string  `json:"name"`
	Provider    *string  `json:"provider"`
	BaseURL     *string  `json:"base_url"`
	APIKey      *string  `json:"api_key"`
	Model       *string  `json:"model"`
	MaxTokens   *int     `json:"max_tokens"`
	Temperature *float64 `json:"temperature"`
	Timeout     *int     `json:"timeout"`
	Purpose     *string  `json:"purpose"`
	IsActive    *bool    `json:"is_active"`
}

// LLMChatRequest 对话请求
type LLMChatRequest struct {
	Messages       []llm.ChatMessage `json:"messages"`
	Purpose        string            `json:"purpose"`
	ModelID        *int64            `json:"model_id"`
	ResponseFormat string            `json:"response_format"` // text / json
	MaxTokens      *int              `json:"max_tokens"`
	Temperature    *float32          `json:"temperature"`
}

// LLMChatResponse 对话响应
type LLMChatResponse struct {
	Content    string `json:"content"`
	Model      string `json:"model"`
	Provider   string `json:"provider"`
	Usage      LLMUsage `json:"usage"`
	DurationMs int    `json:"duration_ms"`
}

// LLMUsage Token用量
type LLMUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// LLMServiceImpl LLM服务实现
type LLMServiceImpl struct {
	modelRepo repository.LLMModelRepository
	logRepo   repository.LLMCallLogRepository
	client    *llm.Client
	cfg       config.LLMConfig
}

// NewLLMService 创建LLM服务
func NewLLMService(modelRepo repository.LLMModelRepository, logRepo repository.LLMCallLogRepository, client *llm.Client, cfg config.LLMConfig) *LLMServiceImpl {
	return &LLMServiceImpl{
		modelRepo: modelRepo,
		logRepo:   logRepo,
		client:    client,
		cfg:       cfg,
	}
}

// ======= 模型配置 CRUD =======

func (s *LLMServiceImpl) CreateModel(ctx context.Context, req LLMModelCreate) (*model.LLMModel, error) {
	if req.Name == "" {
		return nil, errors.New("名称不能为空")
	}
	if req.Provider == "" {
		return nil, errors.New("服务商不能为空")
	}
	if req.BaseURL == "" {
		return nil, errors.New("API端点不能为空")
	}
	if req.APIKey == "" {
		return nil, errors.New("API密钥不能为空")
	}
	if req.Model == "" {
		return nil, errors.New("模型标识不能为空")
	}
	if req.MaxTokens <= 0 {
		req.MaxTokens = 4096
	}
	if req.Temperature <= 0 {
		req.Temperature = 0.7
	}
	if req.Timeout <= 0 {
		req.Timeout = 60
	}
	if req.Purpose == "" {
		req.Purpose = "default"
	}

	now := time.Now()
	m := &model.LLMModel{
		Name:        req.Name,
		Provider:    req.Provider,
		BaseURL:     req.BaseURL,
		APIKey:      req.APIKey,
		Model:       req.Model,
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
		Timeout:     req.Timeout,
		Purpose:     req.Purpose,
		IsActive:    true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.modelRepo.Create(ctx, m); err != nil {
		return nil, err
	}
	return m, nil
}

func (s *LLMServiceImpl) ListModels(ctx context.Context, query repository.LLMModelListQuery) ([]model.LLMModel, int64, error) {
	return s.modelRepo.List(ctx, query)
}

func (s *LLMServiceImpl) GetModel(ctx context.Context, id int64) (*model.LLMModel, error) {
	m, err := s.modelRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("模型配置不存在")
		}
		return nil, err
	}
	return m, nil
}

func (s *LLMServiceImpl) UpdateModel(ctx context.Context, id int64, req LLMModelUpdate) (*model.LLMModel, error) {
	_, err := s.modelRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("模型配置不存在")
		}
		return nil, err
	}

	updates := map[string]interface{}{}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Provider != nil {
		updates["provider"] = *req.Provider
	}
	if req.BaseURL != nil {
		updates["base_url"] = *req.BaseURL
	}
	if req.APIKey != nil {
		updates["api_key"] = *req.APIKey
	}
	if req.Model != nil {
		updates["model"] = *req.Model
	}
	if req.MaxTokens != nil {
		updates["max_tokens"] = *req.MaxTokens
	}
	if req.Temperature != nil {
		updates["temperature"] = *req.Temperature
	}
	if req.Timeout != nil {
		updates["timeout"] = *req.Timeout
	}
	if req.Purpose != nil {
		updates["purpose"] = *req.Purpose
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}
	if len(updates) == 0 {
		return s.modelRepo.FindByID(ctx, id)
	}

	updates["updated_at"] = time.Now()
	if err := s.modelRepo.Update(ctx, id, updates); err != nil {
		return nil, err
	}
	return s.modelRepo.FindByID(ctx, id)
}

func (s *LLMServiceImpl) DeleteModel(ctx context.Context, id int64) error {
	_, err := s.modelRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("模型配置不存在")
		}
		return err
	}
	now := time.Now()
	return s.modelRepo.Update(ctx, id, map[string]interface{}{
		"deleted_at": &now,
		"updated_at": now,
	})
}

// ======= 对话 =======

func (s *LLMServiceImpl) Chat(ctx context.Context, userID int64, req LLMChatRequest) (*LLMChatResponse, error) {
	if len(req.Messages) == 0 {
		return nil, errors.New("messages不能为空")
	}
	if req.Purpose == "" {
		req.Purpose = "default"
	}

	// 确定模型配置：model_id > purpose > config.yaml
	var (
		baseURL     = s.cfg.Default.BaseURL
		apiKey      = s.cfg.Default.APIKey
		modelName   = s.cfg.Default.Model
		provider    = "config"
		maxTokens   = s.cfg.Default.MaxTokens
		temperature = s.cfg.Default.Temperature
		dbModelID   *int64
	)

	if req.ModelID != nil {
		// 通过 model_id 指定模型
		dbModel, err := s.modelRepo.FindByID(ctx, *req.ModelID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errors.New("指定的模型配置不存在")
			}
			return nil, err
		}
		baseURL = dbModel.BaseURL
		apiKey = dbModel.APIKey
		modelName = dbModel.Model
		provider = dbModel.Provider
		maxTokens = dbModel.MaxTokens
		temperature = float32(dbModel.Temperature)
		dbModelID = &dbModel.ID
	} else {
		// 通过 purpose 查找数据库模型
		dbModel, err := s.modelRepo.FindActiveByPurpose(ctx, req.Purpose)
		if err == nil && dbModel != nil {
			baseURL = dbModel.BaseURL
			apiKey = dbModel.APIKey
			modelName = dbModel.Model
			provider = dbModel.Provider
			maxTokens = dbModel.MaxTokens
			temperature = float32(dbModel.Temperature)
			dbModelID = &dbModel.ID
		}
		// 查找失败则使用 config.yaml 默认配置
	}

	if apiKey == "" {
		return nil, errors.New("无可用模型配置")
	}

	// 请求参数覆盖
	if req.MaxTokens != nil {
		maxTokens = *req.MaxTokens
	}
	if req.Temperature != nil {
		temperature = *req.Temperature
	}

	// 构建调用选项
	opts := []llm.ChatOption{
		llm.WithEndpoint(baseURL, apiKey),
		llm.WithModel(modelName),
		llm.WithMaxTokens(maxTokens),
		llm.WithTemperature(temperature),
	}
	if req.ResponseFormat == "json" {
		opts = append(opts, llm.WithJSONMode())
	}

	// 调用 LLM
	result, err := s.client.ChatCompletion(ctx, req.Messages, opts...)

	// 记录日志
	logStatus := int16(1)
	var errMsg *string
	durationMs := 0
	promptTokens := 0
	completionTokens := 0
	totalTokens := 0

	if err != nil {
		logStatus = 2
		msg := err.Error()
		errMsg = &msg
		if msg == "调用超时" {
			logStatus = 3
		}
	} else {
		durationMs = result.DurationMs
		promptTokens = result.PromptTokens
		completionTokens = result.CompletionTokens
		totalTokens = result.TotalTokens
	}

	callLog := &model.LLMCallLog{
		UserID:           userID,
		ModelID:          dbModelID,
		Provider:         provider,
		Model:            modelName,
		Purpose:          req.Purpose,
		PromptTokens:     promptTokens,
		CompletionTokens: completionTokens,
		TotalTokens:      totalTokens,
		DurationMs:       durationMs,
		Status:           logStatus,
		ErrorMessage:     errMsg,
		CreatedAt:        time.Now(),
	}
	if logErr := s.logRepo.Create(ctx, callLog); logErr != nil {
		log.Errorf("记录LLM调用日志失败: %v", logErr)
	}

	if err != nil {
		return nil, err
	}

	return &LLMChatResponse{
		Content:  result.Content,
		Model:    modelName,
		Provider: provider,
		Usage: LLMUsage{
			PromptTokens:     result.PromptTokens,
			CompletionTokens: result.CompletionTokens,
			TotalTokens:      result.TotalTokens,
		},
		DurationMs: result.DurationMs,
	}, nil
}

// ======= 日志查询 =======

func (s *LLMServiceImpl) ListLogs(ctx context.Context, query repository.LLMCallLogListQuery) ([]model.LLMCallLog, int64, error) {
	return s.logRepo.List(ctx, query)
}

func (s *LLMServiceImpl) LogStats(ctx context.Context, query repository.LLMCallLogStatsQuery) (*repository.LLMCallLogStats, []repository.LLMCallLogGroupStats, error) {
	return s.logRepo.Stats(ctx, query)
}
