package model

import (
	"time"

	"gorm.io/datatypes"
)

// Project 项目表
type Project struct {
	ID               int64          `gorm:"primaryKey" json:"id"`
	OwnerUserID      int64          `json:"owner_user_id"`
	Name             string         `gorm:"size:128" json:"name"`
	NarrativeMode    int16          `gorm:"default:1" json:"narrative_mode"`
	CoverResourceID  *int64         `json:"cover_resource_id"`
	VideoAspectRatio string         `gorm:"size:16" json:"video_aspect_ratio"`
	StyleRef         string         `gorm:"size:256" json:"style_ref"`
	Status           int16          `gorm:"default:1" json:"status"`
	ExtraData        datatypes.JSON `gorm:"type:jsonb" json:"extra_data"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        *time.Time     `json:"deleted_at"`
}
