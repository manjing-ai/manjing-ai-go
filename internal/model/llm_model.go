package model

import (
	"time"

	"gorm.io/datatypes"
)

// LLMModel LLM模型配置表
type LLMModel struct {
	ID          int64          `gorm:"primaryKey" json:"id"`
	Name        string         `gorm:"size:64" json:"name"`               // 模型显示名称
	Provider    string         `gorm:"size:32" json:"provider"`           // 服务商标识
	BaseURL     string         `gorm:"size:256" json:"base_url"`          // API端点URL
	APIKey      string         `gorm:"size:256" json:"-"`                 // API密钥（JSON序列化时不输出）
	Model       string         `gorm:"size:64" json:"model"`              // 模型标识
	MaxTokens   int            `gorm:"default:4096" json:"max_tokens"`    // 最大输出Token数
	Temperature float64        `gorm:"default:0.70" json:"temperature"`   // 温度参数
	Timeout     int            `gorm:"default:60" json:"timeout"`         // 超时时间（秒）
	Purpose     string         `gorm:"size:32;default:default" json:"purpose"` // 用途
	IsActive    bool           `gorm:"default:true" json:"is_active"`     // 是否启用
	ExtraConfig datatypes.JSON `gorm:"type:jsonb" json:"extra_config"`    // 扩展配置
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   *time.Time     `json:"deleted_at"`
}

// MaskedAPIKey 返回脱敏后的API Key
func (m *LLMModel) MaskedAPIKey() string {
	if len(m.APIKey) <= 6 {
		return "sk-***"
	}
	return m.APIKey[:3] + "***" + m.APIKey[len(m.APIKey)-3:]
}
