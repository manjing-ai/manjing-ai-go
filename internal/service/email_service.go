package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"regexp"
	"strings"
	"time"

	"manjing-ai-go/config"
	"manjing-ai-go/pkg/email"
	"manjing-ai-go/pkg/logger"
	redisclient "manjing-ai-go/pkg/redis"
)

// EmailService 邮件服务
type EmailService interface {
	SendVerifyCode(ctx context.Context, req EmailSendReq) (EmailSendResp, error)
}

// EmailSendReq 发送验证码请求
type EmailSendReq struct {
	Email string
	Scene string
}

// EmailSendResp 发送验证码响应
type EmailSendResp struct {
	RequestID     string
	ExpireSeconds int
	NextSendAfter int
}

// EmailServiceImpl 实现
type EmailServiceImpl struct {
	cfg   config.EmailConfig
	rdb   *redisclient.Client
	email email.Client
}

// NewEmailService 创建服务
func NewEmailService(cfg config.EmailConfig, rdb *redisclient.Client, client email.Client) *EmailServiceImpl {
	return &EmailServiceImpl{cfg: cfg, rdb: rdb, email: client}
}

func (s *EmailServiceImpl) SendVerifyCode(ctx context.Context, req EmailSendReq) (EmailSendResp, error) {
	if !isValidEmail(req.Email) {
		return EmailSendResp{}, errors.New("邮箱格式不正确")
	}
	scene := strings.TrimSpace(req.Scene)
	if scene == "" {
		scene = "register"
	}
	sceneCfg, ok := s.cfg.Scenes[scene]
	if !ok {
		return EmailSendResp{}, errors.New("场景未配置")
	}
	if s.rdb == nil {
		return EmailSendResp{}, errors.New("验证码服务不可用")
	}

	interval := s.cfg.RateLimit.IntervalSeconds
	if interval <= 0 {
		interval = 60
	}
	rateKey := emailRateKey(scene, req.Email)
	exists, err := s.rdb.RDB.Exists(ctx, rateKey).Result()
	if err != nil {
		return EmailSendResp{}, err
	}
	if exists == 1 {
		return EmailSendResp{}, errors.New("发送过于频繁")
	}

	expire := sceneCfg.TTLSeconds
	if expire <= 0 {
		expire = s.cfg.Code.TTLSeconds
	}
	if expire <= 0 {
		expire = 300
	}

	codeLen := s.cfg.Code.Length
	if codeLen <= 0 {
		codeLen = 6
	}
	code, err := generateCode(codeLen)
	if err != nil {
		return EmailSendResp{}, err
	}

	tpl := s.pickTemplate(sceneCfg.TemplateCode)
	if tpl == "" {
		return EmailSendResp{}, errors.New("验证码模板未配置")
	}
	subject := sceneCfg.Subject
	if subject == "" {
		subject = s.pickSubject(sceneCfg.TemplateCode)
	}
	if subject == "" {
		subject = "验证码"
	}
	body, err := email.Render(tpl, map[string]interface{}{
		"code":           code,
		"expire_seconds": expire,
		"expire_minutes": expire / 60,
		"scene":          scene,
	})
	if err != nil {
		return EmailSendResp{}, err
	}

	if err := s.email.Send(ctx, req.Email, subject, body); err != nil {
		logger.L().WithError(err).Warn("email send failed")
		return EmailSendResp{}, errors.New("邮件发送失败")
	}

	codeKey := emailCodeKey(scene, req.Email)
	if err := s.rdb.RDB.Set(ctx, codeKey, code, time.Duration(expire)*time.Second).Err(); err != nil {
		return EmailSendResp{}, err
	}
	if err := s.rdb.RDB.Set(ctx, rateKey, 1, time.Duration(interval)*time.Second).Err(); err != nil {
		return EmailSendResp{}, err
	}

	return EmailSendResp{
		RequestID:     buildRequestID(),
		ExpireSeconds: expire,
		NextSendAfter: interval,
	}, nil
}

func (s *EmailServiceImpl) pickTemplate(code string) string {
	if code != "" {
		if tpl, ok := s.cfg.Templates[code]; ok {
			return tpl
		}
	}
	if s.cfg.Templates == nil {
		return ""
	}
	if tpl, ok := s.cfg.Templates["default"]; ok {
		return tpl
	}
	for _, v := range s.cfg.Templates {
		return v
	}
	return ""
}

func (s *EmailServiceImpl) pickSubject(code string) string {
	if code != "" {
		if sub, ok := s.cfg.Subjects[code]; ok {
			return sub
		}
	}
	if s.cfg.Subjects == nil {
		return ""
	}
	if sub, ok := s.cfg.Subjects["default"]; ok {
		return sub
	}
	for _, v := range s.cfg.Subjects {
		return v
	}
	return ""
}

func emailCodeKey(scene, emailAddr string) string {
	return fmt.Sprintf("email:code:%s:%s", scene, emailAddr)
}

func emailRateKey(scene, emailAddr string) string {
	return fmt.Sprintf("email:rate:%s:%s", scene, emailAddr)
}

func generateCode(length int) (string, error) {
	if length <= 0 {
		return "", errors.New("验证码长度非法")
	}
	var b strings.Builder
	for i := 0; i < length; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", err
		}
		b.WriteByte(byte('0' + n.Int64()))
	}
	return b.String(), nil
}

func buildRequestID() string {
	ts := time.Now().Format("20060102_150405")
	n, _ := rand.Int(rand.Reader, big.NewInt(10000))
	return fmt.Sprintf("email_req_%s_%04d", ts, n.Int64())
}

var emailRegex = regexp.MustCompile(`^[A-Za-z0-9._%+\-]+@[A-Za-z0-9.\-]+\.[A-Za-z]{2,}$`)

func isValidEmail(v string) bool {
	return emailRegex.MatchString(strings.TrimSpace(v))
}
