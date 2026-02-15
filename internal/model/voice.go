package model

import (
	"time"

	"gorm.io/datatypes"
)

// Voice 声音表
type Voice struct {
	ID          int64          `gorm:"primaryKey" json:"id"`
	Name        string         `gorm:"size:64" json:"name"`            // 声音名称
	AgeGroup    int16          `json:"age_group"`                      // 年龄段：1儿童/2少年/3青年/4中年/5老年
	Gender      int16          `json:"gender"`                         // 性别：1男/2女
	Dialect     int16          `gorm:"default:1" json:"dialect"`       // 方言口音：1标准普通话/2东北话/3四川话/4粤语/5台湾腔/6港式普通话/7外国口音
	Tone        int16          `gorm:"default:1" json:"tone"`          // 音色：1标准/2清亮/3浑厚/4沙哑/5柔和/6尖细/7气声/8鼻音/9金属
	SampleURL   string         `gorm:"size:512" json:"sample_url"`     // 试听音频URL
	Type        int16          `gorm:"default:1" json:"type"`          // 类型：1官方/2用户
	OwnerUserID *int64         `json:"owner_user_id"`                  // 所属用户ID（type=2时必填）
	ExtraData   datatypes.JSON `gorm:"type:jsonb" json:"extra_data"`   // 扩展字段
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   *time.Time     `json:"deleted_at"`
}
