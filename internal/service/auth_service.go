package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"manjing-ai-go/config"
	"manjing-ai-go/internal/model"
	"manjing-ai-go/internal/repository"
	"manjing-ai-go/pkg/jwtutil"
	redisclient "manjing-ai-go/pkg/redis"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// AuthService 用户认证服务
type AuthService interface {
	Register(ctx context.Context, email, phone, username, password string) (interface{}, error)
	Login(ctx context.Context, account, password string) (interface{}, error)
	Profile(ctx context.Context, userID int64) (interface{}, error)
	ChangePassword(ctx context.Context, userID int64, oldPassword, newPassword string) error
	Logout(ctx context.Context, userID int64, token string) error
	UpdateStatus(ctx context.Context, userID int64, status int16) error
	UpdateAvatar(ctx context.Context, userID int64, avatarURL string) error
}

// AuthServiceImpl 实现
type AuthServiceImpl struct {
	repo repository.UserRepository
	jwt  config.JWTConfig
	rdb  *redisclient.Client
}

// NewAuthService 创建服务
func NewAuthService(repo repository.UserRepository, jwtCfg config.JWTConfig, rdb *redisclient.Client) *AuthServiceImpl {
	return &AuthServiceImpl{repo: repo, jwt: jwtCfg, rdb: rdb}
}

func (s *AuthServiceImpl) Register(ctx context.Context, email, phone, username, password string) (interface{}, error) {
	if password == "" {
		return nil, errors.New("密码不能为空")
	}
	if email == "" && phone == "" {
		return nil, errors.New("邮箱或手机号不能为空")
	}

	if email != "" {
		if _, err := s.repo.FindByEmail(ctx, email); err == nil {
			return nil, errors.New("邮箱已存在")
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
	}
	if phone != "" {
		if _, err := s.repo.FindByPhone(ctx, phone); err == nil {
			return nil, errors.New("手机号已存在")
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		Username:     username,
		Email:        email,
		Phone:        phone,
		PasswordHash: string(hash),
		Status:       1,
		Role:         "user",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}

	token, err := jwtutil.Generate(user.ID, s.jwt)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"token": token,
		"user": map[string]interface{}{
			"id":       user.ID,
			"username": username,
			"email":    email,
		},
	}, nil
}

func (s *AuthServiceImpl) Login(ctx context.Context, account, password string) (interface{}, error) {
	if account == "" || password == "" {
		return nil, errors.New("账号或密码不能为空")
	}
	var user *model.User
	var err error
	if strings.Contains(account, "@") {
		user, err = s.repo.FindByEmail(ctx, account)
	} else {
		user, err = s.repo.FindByPhone(ctx, account)
	}
	if err != nil {
		return nil, errors.New("账号或密码错误")
	}
	if user.Status != 1 {
		return nil, errors.New("账号被禁用")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, errors.New("账号或密码错误")
	}

	_ = s.repo.UpdateLastLogin(ctx, user.ID, time.Now())

	token, err := jwtutil.Generate(user.ID, s.jwt)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"token": token,
		"user": map[string]interface{}{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
		},
	}, nil
}

func (s *AuthServiceImpl) Profile(ctx context.Context, userID int64) (interface{}, error) {
	if userID == 0 {
		return nil, errors.New("未授权")
	}
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"id":         user.ID,
		"username":   user.Username,
		"email":      user.Email,
		"phone":      user.Phone,
		"avatar_url": user.AvatarURL,
		"role":       user.Role,
	}, nil
}

func (s *AuthServiceImpl) ChangePassword(ctx context.Context, userID int64, oldPassword, newPassword string) error {
	if userID == 0 {
		return errors.New("未授权")
	}
	if newPassword == "" {
		return errors.New("新密码不能为空")
	}
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword)); err != nil {
		return errors.New("旧密码错误")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return s.repo.UpdatePassword(ctx, userID, string(hash))
}

func (s *AuthServiceImpl) Logout(ctx context.Context, userID int64, token string) error {
	if userID == 0 {
		return errors.New("未授权")
	}
	if s.rdb == nil {
		return nil
	}
	if token == "" {
		return nil
	}
	claims, err := jwtutil.Parse(token, s.jwt)
	if err != nil || claims.ExpiresAt == nil {
		return nil
	}
	ttl := time.Until(claims.ExpiresAt.Time)
	if ttl <= 0 {
		return nil
	}
	return s.rdb.SetTokenBlacklisted(ctx, token, ttl)
}

func (s *AuthServiceImpl) UpdateStatus(ctx context.Context, userID int64, status int16) error {
	if userID == 0 {
		return errors.New("参数错误")
	}
	if status != 0 && status != 1 {
		return errors.New("状态非法")
	}
	return s.repo.UpdateStatus(ctx, userID, status)
}

func (s *AuthServiceImpl) UpdateAvatar(ctx context.Context, userID int64, avatarURL string) error {
	if userID == 0 {
		return errors.New("参数错误")
	}
	if avatarURL == "" {
		return errors.New("头像不能为空")
	}
	return s.repo.UpdateAvatar(ctx, userID, avatarURL)
}
