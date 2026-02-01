package repository

import (
	"context"
	"time"

	"manjing-ai-go/internal/model"

	"gorm.io/gorm"
)

// UserRepository 用户数据访问接口
type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	FindByEmail(ctx context.Context, email string) (*model.User, error)
	FindByPhone(ctx context.Context, phone string) (*model.User, error)
	FindByID(ctx context.Context, id int64) (*model.User, error)
	UpdatePassword(ctx context.Context, id int64, passwordHash string) error
	UpdateLastLogin(ctx context.Context, id int64, lastLoginAt time.Time) error
	UpdateStatus(ctx context.Context, id int64, status int16) error
	UpdateAvatar(ctx context.Context, id int64, avatarURL string) error
}

// UserRepo 实现
type UserRepo struct {
	db *gorm.DB
}

// NewUserRepo 创建仓库
func NewUserRepo(db *gorm.DB) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) Create(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *UserRepo) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepo) FindByPhone(ctx context.Context, phone string) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Where("phone = ?", phone).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepo) FindByID(ctx context.Context, id int64) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepo) UpdatePassword(ctx context.Context, id int64, passwordHash string) error {
	return r.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", id).Update("password_hash", passwordHash).Error
}

func (r *UserRepo) UpdateLastLogin(ctx context.Context, id int64, lastLoginAt time.Time) error {
	return r.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", id).Update("last_login_at", lastLoginAt).Error
}

func (r *UserRepo) UpdateStatus(ctx context.Context, id int64, status int16) error {
	return r.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", id).Update("status", status).Error
}

func (r *UserRepo) UpdateAvatar(ctx context.Context, id int64, avatarURL string) error {
	return r.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", id).Update("avatar_url", avatarURL).Error
}
