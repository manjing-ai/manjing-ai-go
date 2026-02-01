package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"manjing-ai-go/internal/service"
)

// AuthHandler 用户认证相关处理器
type AuthHandler struct {
	svc service.AuthService
}

// NewAuthHandler 创建处理器
func NewAuthHandler(svc service.AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

// RegisterReq 注册请求
type RegisterReq struct {
	Email    string `json:"email"`    // 邮箱（可选）
	Phone    string `json:"phone"`    // 手机号（可选）
	Password string `json:"password"` // 密码（必填）
	Username string `json:"username"` // 用户名（可选）
}

// LoginReq 登录请求
type LoginReq struct {
	Account  string `json:"account"`  // 账号（邮箱/手机号）
	Password string `json:"password"` // 密码
}

// PasswordReq 修改密码请求
type PasswordReq struct {
	OldPassword string `json:"old_password"` // 旧密码
	NewPassword string `json:"new_password"` // 新密码
}

// StatusReq 更新状态请求
type StatusReq struct {
	Status int16 `json:"status"` // 状态（0禁用/1正常）
}

// AvatarReq 更新头像请求
type AvatarReq struct {
	AvatarURL string `json:"avatar_url"` // 头像URL
}

// Register 用户注册
// @Summary 用户注册
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body RegisterReq true "注册信息"
// @Success 200 {object} Resp
// @Router /api/v1/auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterReq
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 10001, "参数错误")
		return
	}
	resp, err := h.svc.Register(c.Request.Context(), req.Email, req.Phone, req.Username, req.Password)
	if err != nil {
		fail(c, 10001, err.Error())
		return
	}
	ok(c, resp)
}

// Login 用户登录
// @Summary 用户登录
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body LoginReq true "登录信息"
// @Success 200 {object} Resp
// @Router /api/v1/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 10001, "参数错误")
		return
	}
	resp, err := h.svc.Login(c.Request.Context(), req.Account, req.Password)
	if err != nil {
		fail(c, 10003, err.Error())
		return
	}
	ok(c, resp)
}

// Profile 获取用户信息
// @Summary 获取用户信息
// @Tags Auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} Resp
// @Router /api/v1/auth/profile [get]
func (h *AuthHandler) Profile(c *gin.Context) {
	userID := c.GetInt64("user_id")
	resp, err := h.svc.Profile(c.Request.Context(), userID)
	if err != nil {
		fail(c, 10004, err.Error())
		return
	}
	ok(c, resp)
}

// ChangePassword 修改密码
// @Summary 修改密码
// @Tags Auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body PasswordReq true "修改密码"
// @Success 200 {object} Resp
// @Router /api/v1/auth/password [put]
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	var req PasswordReq
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 10001, "参数错误")
		return
	}
	userID := c.GetInt64("user_id")
	if err := h.svc.ChangePassword(c.Request.Context(), userID, req.OldPassword, req.NewPassword); err != nil {
		fail(c, 10001, err.Error())
		return
	}
	ok(c, map[string]interface{}{})
}

// Logout 退出登录
// @Summary 退出登录
// @Tags Auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} Resp
// @Router /api/v1/auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	token := c.GetString("token")
	if err := h.svc.Logout(c.Request.Context(), c.GetInt64("user_id"), token); err != nil {
		c.JSON(http.StatusOK, Resp{Code: 10001, Message: err.Error(), Data: map[string]interface{}{}})
		return
	}
	ok(c, map[string]interface{}{})
}

// UpdateStatus 更新用户状态
// @Summary 更新用户状态
// @Tags User
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "用户ID"
// @Param body body StatusReq true "状态"
// @Success 200 {object} Resp
// @Router /api/v1/users/{id}/status [put]
func (h *AuthHandler) UpdateStatus(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		fail(c, 10001, "参数错误")
		return
	}
	var req StatusReq
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 10001, "参数错误")
		return
	}
	if err := h.svc.UpdateStatus(c.Request.Context(), id, req.Status); err != nil {
		fail(c, 10001, err.Error())
		return
	}
	ok(c, map[string]interface{}{})
}

// UpdateAvatar 更新用户头像
// @Summary 更新用户头像
// @Tags User
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "用户ID"
// @Param body body AvatarReq true "头像"
// @Success 200 {object} Resp
// @Router /api/v1/users/{id}/avatar [put]
func (h *AuthHandler) UpdateAvatar(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		fail(c, 10001, "参数错误")
		return
	}
	var req AvatarReq
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 10001, "参数错误")
		return
	}
	if err := h.svc.UpdateAvatar(c.Request.Context(), id, req.AvatarURL); err != nil {
		fail(c, 10001, err.Error())
		return
	}
	ok(c, map[string]interface{}{})
}
