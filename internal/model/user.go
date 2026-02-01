package model

import "time"

// User 用户表结构
type User struct {
	ID           int64     `gorm:"primaryKey" json:"id"`
	Username     string    `gorm:"size:64" json:"username"`
	Email        string    `gorm:"size:128" json:"email"`
	Phone        string    `gorm:"size:32" json:"phone"`
	PasswordHash string    `gorm:"size:256" json:"-"`
	AvatarURL    string    `gorm:"size:512" json:"avatar_url"`
	Status       int16     `gorm:"default:1" json:"status"`
	Role         string    `gorm:"size:32" json:"role"`
	LastLoginAt  time.Time `json:"last_login_at"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
