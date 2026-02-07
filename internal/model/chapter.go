package model

import (
	"time"

	"gorm.io/datatypes"
)

// Chapter 章节表
type Chapter struct {
	ID         int64          `gorm:"primaryKey" json:"id"`
	ProjectID  int64          `json:"project_id"`
	Name       string         `gorm:"size:128" json:"name"`
	Content    string         `json:"content"`
	Summary    string         `gorm:"size:256" json:"summary"`
	OrderIndex int            `gorm:"default:0" json:"order_index"`
	Status     int16          `gorm:"default:1" json:"status"`
	ExtraData  datatypes.JSON `gorm:"type:jsonb" json:"extra_data"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  *time.Time     `json:"deleted_at"`
}
