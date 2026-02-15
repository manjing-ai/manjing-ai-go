package repository

import (
	"context"
	"time"

	"manjing-ai-go/internal/model"

	"gorm.io/gorm"
)

// LLMCallLogRepository LLM调用日志数据访问接口
type LLMCallLogRepository interface {
	Create(ctx context.Context, log *model.LLMCallLog) error
	List(ctx context.Context, query LLMCallLogListQuery) ([]model.LLMCallLog, int64, error)
	Stats(ctx context.Context, query LLMCallLogStatsQuery) (*LLMCallLogStats, []LLMCallLogGroupStats, error)
}

// LLMCallLogListQuery 调用日志列表查询参数
type LLMCallLogListQuery struct {
	Page      int
	PageSize  int
	Purpose   string
	Provider  string
	Status    int
	StartTime *time.Time
	EndTime   *time.Time
	Sort      string
}

// LLMCallLogStatsQuery 用量统计查询参数
type LLMCallLogStatsQuery struct {
	StartTime *time.Time
	EndTime   *time.Time
	GroupBy   string // provider / purpose / model
}

// LLMCallLogStats 用量统计结果
type LLMCallLogStats struct {
	TotalCalls            int64 `json:"total_calls"`
	SuccessCalls          int64 `json:"success_calls"`
	FailedCalls           int64 `json:"failed_calls"`
	TotalTokens           int64 `json:"total_tokens"`
	TotalPromptTokens     int64 `json:"total_prompt_tokens"`
	TotalCompletionTokens int64 `json:"total_completion_tokens"`
	AvgDurationMs         int64 `json:"avg_duration_ms"`
}

// LLMCallLogGroupStats 分组统计结果
type LLMCallLogGroupStats struct {
	Key           string `json:"key" gorm:"column:group_key"`
	CallCount     int64  `json:"call_count"`
	TotalTokens   int64  `json:"total_tokens"`
	AvgDurationMs int64  `json:"avg_duration_ms"`
}

// LLMCallLogRepo 调用日志仓库实现
type LLMCallLogRepo struct {
	db *gorm.DB
}

// NewLLMCallLogRepo 创建调用日志仓库
func NewLLMCallLogRepo(db *gorm.DB) *LLMCallLogRepo {
	return &LLMCallLogRepo{db: db}
}

func (r *LLMCallLogRepo) Create(ctx context.Context, log *model.LLMCallLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *LLMCallLogRepo) List(ctx context.Context, query LLMCallLogListQuery) ([]model.LLMCallLog, int64, error) {
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 20
	}
	if query.PageSize > 100 {
		query.PageSize = 100
	}

	db := r.db.WithContext(ctx).Model(&model.LLMCallLog{})

	if query.Purpose != "" {
		db = db.Where("purpose = ?", query.Purpose)
	}
	if query.Provider != "" {
		db = db.Where("provider = ?", query.Provider)
	}
	if query.Status > 0 {
		db = db.Where("status = ?", query.Status)
	}
	if query.StartTime != nil {
		db = db.Where("created_at >= ?", *query.StartTime)
	}
	if query.EndTime != nil {
		db = db.Where("created_at <= ?", *query.EndTime)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	switch query.Sort {
	case "created_at":
		db = db.Order("created_at ASC")
	case "-duration_ms":
		db = db.Order("duration_ms DESC")
	default:
		db = db.Order("created_at DESC")
	}

	var items []model.LLMCallLog
	err := db.Offset((query.Page - 1) * query.PageSize).Limit(query.PageSize).Find(&items).Error
	return items, total, err
}

func (r *LLMCallLogRepo) Stats(ctx context.Context, query LLMCallLogStatsQuery) (*LLMCallLogStats, []LLMCallLogGroupStats, error) {
	db := r.db.WithContext(ctx).Model(&model.LLMCallLog{})

	if query.StartTime != nil {
		db = db.Where("created_at >= ?", *query.StartTime)
	}
	if query.EndTime != nil {
		db = db.Where("created_at <= ?", *query.EndTime)
	}

	// 总体统计
	var stats LLMCallLogStats
	err := db.Select(`
		COUNT(*) AS total_calls,
		COUNT(CASE WHEN status = 1 THEN 1 END) AS success_calls,
		COUNT(CASE WHEN status != 1 THEN 1 END) AS failed_calls,
		COALESCE(SUM(total_tokens), 0) AS total_tokens,
		COALESCE(SUM(prompt_tokens), 0) AS total_prompt_tokens,
		COALESCE(SUM(completion_tokens), 0) AS total_completion_tokens,
		COALESCE(AVG(duration_ms), 0) AS avg_duration_ms
	`).Scan(&stats).Error
	if err != nil {
		return nil, nil, err
	}

	// 分组统计
	groupBy := query.GroupBy
	if groupBy == "" {
		groupBy = "provider"
	}

	db2 := r.db.WithContext(ctx).Model(&model.LLMCallLog{})
	if query.StartTime != nil {
		db2 = db2.Where("created_at >= ?", *query.StartTime)
	}
	if query.EndTime != nil {
		db2 = db2.Where("created_at <= ?", *query.EndTime)
	}

	var groups []LLMCallLogGroupStats
	err = db2.Select(groupBy + ` AS group_key,
		COUNT(*) AS call_count,
		COALESCE(SUM(total_tokens), 0) AS total_tokens,
		COALESCE(AVG(duration_ms), 0) AS avg_duration_ms
	`).Group(groupBy).Scan(&groups).Error
	if err != nil {
		return nil, nil, err
	}

	return &stats, groups, nil
}
