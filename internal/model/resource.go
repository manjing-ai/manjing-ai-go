package model

import (
	"time"

	"gorm.io/datatypes"
)

// Resource 用户资源表
type Resource struct {
	ID        int64          `gorm:"primaryKey" json:"id"`
	UserID    int64          `json:"user_id"`
	Name      string         `gorm:"size:255" json:"name"`
	Type      string         `gorm:"size:32" json:"type"`
	Category  string         `gorm:"size:32" json:"category"`
	ObjectKey string         `gorm:"size:512" json:"object_key"`
	FileName  string         `gorm:"size:255" json:"file_name"`
	FileExt   string         `gorm:"size:16" json:"file_ext"`
	MimeType  string         `gorm:"size:64" json:"mime_type"`
	Width     int            `json:"width"`
	Height    int            `json:"height"`
	Aspect    string         `gorm:"size:16" json:"aspect"`
	SizeBytes int64          `json:"size_bytes"`
	Status    string         `gorm:"size:16" json:"status"`
	ExtraData datatypes.JSON `gorm:"type:jsonb" json:"extra_data"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt *time.Time     `json:"deleted_at"`
}
