package model

import "time"

// LLMCallLog LLM调用日志表
type LLMCallLog struct {
	ID               int64      `gorm:"primaryKey" json:"id"`
	UserID           int64      `json:"user_id"`                                  // 调用用户ID
	ModelID          *int64     `json:"model_id"`                                 // 关联模型配置ID
	Provider         string     `gorm:"size:32" json:"provider"`                  // 服务商标识
	Model            string     `gorm:"size:64" json:"model"`                     // 实际使用的模型标识
	Purpose          string     `gorm:"size:32" json:"purpose"`                   // 调用用途
	PromptTokens     int        `gorm:"default:0" json:"prompt_tokens"`           // 输入Token数
	CompletionTokens int        `gorm:"default:0" json:"completion_tokens"`       // 输出Token数
	TotalTokens      int        `gorm:"default:0" json:"total_tokens"`            // 总Token数
	DurationMs       int        `gorm:"default:0" json:"duration_ms"`             // 调用耗时（毫秒）
	Status           int16      `gorm:"default:1" json:"status"`                  // 状态：1成功/2失败/3超时
	ErrorMessage     *string    `json:"error_message"`                            // 错误信息
	CreatedAt        time.Time  `json:"created_at"`
}
