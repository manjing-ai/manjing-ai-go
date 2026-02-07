package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"manjing-ai-go/internal/service"
)

// EmailHandler 邮件处理器
type EmailHandler struct {
	svc service.EmailService
}

// NewEmailHandler 创建处理器
func NewEmailHandler(svc service.EmailService) *EmailHandler {
	return &EmailHandler{svc: svc}
}

// SendVerifyCodeReq 发送验证码请求
type SendVerifyCodeReq struct {
	Email string `json:"email"` // 邮箱（必填）
	Scene string `json:"scene"` // 场景（register/reset_password/login）
}

// SendVerifyCode 发送验证码
// @Summary 发送验证码邮件
// @Tags Email
// @Accept json
// @Produce json
// @Param body body SendVerifyCodeReq true "发送验证码"
// @Success 201 {object} Resp
// @Router /v1/emails/verify-codes [post]
func (h *EmailHandler) SendVerifyCode(c *gin.Context) {
	var req SendVerifyCodeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 40001, "参数错误")
		return
	}
	resp, err := h.svc.SendVerifyCode(c.Request.Context(), service.EmailSendReq{
		Email: req.Email,
		Scene: req.Scene,
	})
	if err != nil {
		fail(c, mapEmailErr(err), err.Error())
		return
	}
	c.JSON(http.StatusCreated, Resp{
		Code:    0,
		Message: "success",
		Data: map[string]interface{}{
			"request_id":      resp.RequestID,
			"expire_seconds":  resp.ExpireSeconds,
			"next_send_after": resp.NextSendAfter,
		},
	})
}

func mapEmailErr(err error) int {
	if err == nil {
		return 0
	}
	switch err.Error() {
	case "邮箱格式不正确", "场景未配置", "验证码模板未配置":
		return 40001
	case "发送过于频繁":
		return 42901
	default:
		return 50001
	}
}
