package handler

import (
	"net/http"
	"time"

	"manjing-ai-go/internal/model"
	"manjing-ai-go/internal/repository"
	"manjing-ai-go/internal/service"
	"manjing-ai-go/pkg/llm"

	"github.com/gin-gonic/gin"
)

// LLMHandler LLM处理器
type LLMHandler struct {
	svc service.LLMService
}

// NewLLMHandler 创建LLM处理器
func NewLLMHandler(svc service.LLMService) *LLMHandler {
	return &LLMHandler{svc: svc}
}

// ======= 模型配置 =======

// CreateLLMModelReq 创建模型配置请求
type CreateLLMModelReq struct {
	Name        string  `json:"name"`        // 显示名称（必填）
	Provider    string  `json:"provider"`    // 服务商标识（必填）
	BaseURL     string  `json:"base_url"`    // API端点（必填）
	APIKey      string  `json:"api_key"`     // API密钥（必填）
	Model       string  `json:"model"`       // 模型标识（必填）
	MaxTokens   int     `json:"max_tokens"`  // 最大输出Token
	Temperature float64 `json:"temperature"` // 温度参数
	Timeout     int     `json:"timeout"`     // 超时时间
	Purpose     string  `json:"purpose"`     // 用途
}

// UpdateLLMModelReq 更新模型配置请求
type UpdateLLMModelReq struct {
	Name        *string  `json:"name"`
	Provider    *string  `json:"provider"`
	BaseURL     *string  `json:"base_url"`
	APIKey      *string  `json:"api_key"`
	Model       *string  `json:"model"`
	MaxTokens   *int     `json:"max_tokens"`
	Temperature *float64 `json:"temperature"`
	Timeout     *int     `json:"timeout"`
	Purpose     *string  `json:"purpose"`
	IsActive    *bool    `json:"is_active"`
}

// CreateModel 创建模型配置
// @Summary 创建模型配置
// @Tags LLM
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body CreateLLMModelReq true "创建模型配置"
// @Success 201 {object} Resp
// @Router /v1/llm/models [post]
func (h *LLMHandler) CreateModel(c *gin.Context) {
	var req CreateLLMModelReq
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 40001, "参数错误")
		return
	}
	m, err := h.svc.CreateModel(c.Request.Context(), service.LLMModelCreate{
		Name:        req.Name,
		Provider:    req.Provider,
		BaseURL:     req.BaseURL,
		APIKey:      req.APIKey,
		Model:       req.Model,
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
		Timeout:     req.Timeout,
		Purpose:     req.Purpose,
	})
	if err != nil {
		fail(c, mapLLMErr(err), err.Error())
		return
	}
	c.JSON(http.StatusCreated, Resp{
		Code:    0,
		Message: "success",
		Data:    llmModelToMap(m),
	})
}

// ListModels 模型配置列表
// @Summary 模型配置列表
// @Tags LLM
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码"
// @Param page_size query int false "每页数量"
// @Param provider query string false "服务商"
// @Param purpose query string false "用途"
// @Success 200 {object} Resp
// @Router /v1/llm/models [get]
func (h *LLMHandler) ListModels(c *gin.Context) {
	query := repository.LLMModelListQuery{
		Page:     parseIntDef(c.Query("page"), 1),
		PageSize: parseIntDef(c.Query("page_size"), 20),
		Provider: c.Query("provider"),
		Purpose:  c.Query("purpose"),
	}
	if isActiveStr := c.Query("is_active"); isActiveStr != "" {
		v := isActiveStr == "true"
		query.IsActive = &v
	}
	items, total, err := h.svc.ListModels(c.Request.Context(), query)
	if err != nil {
		fail(c, mapLLMErr(err), err.Error())
		return
	}
	list := make([]map[string]interface{}, 0, len(items))
	for i := range items {
		list = append(list, llmModelToMap(&items[i]))
	}
	ok(c, map[string]interface{}{
		"items": list,
		"pagination": map[string]interface{}{
			"page":        query.Page,
			"page_size":   query.PageSize,
			"total":       total,
			"total_pages": (total + int64(query.PageSize) - 1) / int64(query.PageSize),
		},
	})
}

// ModelDetail 模型配置详情
// @Summary 模型配置详情
// @Tags LLM
// @Produce json
// @Security BearerAuth
// @Param id path int true "模型配置ID"
// @Success 200 {object} Resp
// @Router /v1/llm/models/{id} [get]
func (h *LLMHandler) ModelDetail(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		fail(c, 40001, "参数错误")
		return
	}
	m, err := h.svc.GetModel(c.Request.Context(), id)
	if err != nil {
		fail(c, mapLLMErr(err), err.Error())
		return
	}
	ok(c, llmModelToMap(m))
}

// UpdateModel 更新模型配置
// @Summary 更新模型配置
// @Tags LLM
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "模型配置ID"
// @Param body body UpdateLLMModelReq true "更新模型配置"
// @Success 200 {object} Resp
// @Router /v1/llm/models/{id} [put]
func (h *LLMHandler) UpdateModel(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		fail(c, 40001, "参数错误")
		return
	}
	var req UpdateLLMModelReq
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 40001, "参数错误")
		return
	}
	m, err := h.svc.UpdateModel(c.Request.Context(), id, service.LLMModelUpdate{
		Name:        req.Name,
		Provider:    req.Provider,
		BaseURL:     req.BaseURL,
		APIKey:      req.APIKey,
		Model:       req.Model,
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
		Timeout:     req.Timeout,
		Purpose:     req.Purpose,
		IsActive:    req.IsActive,
	})
	if err != nil {
		fail(c, mapLLMErr(err), err.Error())
		return
	}
	ok(c, llmModelToMap(m))
}

// DeleteModel 删除模型配置
// @Summary 删除模型配置
// @Tags LLM
// @Produce json
// @Security BearerAuth
// @Param id path int true "模型配置ID"
// @Success 204
// @Router /v1/llm/models/{id} [delete]
func (h *LLMHandler) DeleteModel(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		fail(c, 40001, "参数错误")
		return
	}
	if err := h.svc.DeleteModel(c.Request.Context(), id); err != nil {
		fail(c, mapLLMErr(err), err.Error())
		return
	}
	c.Status(http.StatusNoContent)
}

// ======= 对话 =======

// ChatReq 对话请求
type ChatReq struct {
	Messages []llm.ChatMessage `json:"messages"` // 消息列表
	Purpose  string            `json:"purpose"`  // 用途
	ModelID  *int64            `json:"model_id"` // 指定模型配置ID
	ResponseFormat string     `json:"response_format"` // text / json
	MaxTokens      *int       `json:"max_tokens"`
	Temperature    *float32   `json:"temperature"`
}

// Chat 发送对话请求
// @Summary 发送对话请求
// @Tags LLM
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body ChatReq true "对话请求"
// @Success 200 {object} Resp
// @Router /v1/llm/chat [post]
func (h *LLMHandler) Chat(c *gin.Context) {
	userID := c.GetInt64("user_id")
	var req ChatReq
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 40001, "参数错误")
		return
	}
	resp, err := h.svc.Chat(c.Request.Context(), userID, service.LLMChatRequest{
		Messages:       req.Messages,
		Purpose:        req.Purpose,
		ModelID:        req.ModelID,
		ResponseFormat: req.ResponseFormat,
		MaxTokens:      req.MaxTokens,
		Temperature:    req.Temperature,
	})
	if err != nil {
		fail(c, mapLLMErr(err), err.Error())
		return
	}
	ok(c, resp)
}

// ======= 调用日志 =======

// ListLogs 调用日志列表
// @Summary 调用日志列表
// @Tags LLM
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码"
// @Param page_size query int false "每页数量"
// @Param purpose query string false "用途"
// @Param provider query string false "服务商"
// @Param status query int false "状态"
// @Param sort query string false "排序"
// @Success 200 {object} Resp
// @Router /v1/llm/logs [get]
func (h *LLMHandler) ListLogs(c *gin.Context) {
	query := repository.LLMCallLogListQuery{
		Page:     parseIntDef(c.Query("page"), 1),
		PageSize: parseIntDef(c.Query("page_size"), 20),
		Purpose:  c.Query("purpose"),
		Provider: c.Query("provider"),
		Status:   parseIntDef(c.Query("status"), 0),
		Sort:     c.Query("sort"),
	}
	if st := c.Query("start_time"); st != "" {
		if t, err := time.Parse(time.RFC3339, st); err == nil {
			query.StartTime = &t
		}
	}
	if et := c.Query("end_time"); et != "" {
		if t, err := time.Parse(time.RFC3339, et); err == nil {
			query.EndTime = &t
		}
	}
	items, total, err := h.svc.ListLogs(c.Request.Context(), query)
	if err != nil {
		fail(c, mapLLMErr(err), err.Error())
		return
	}
	ok(c, map[string]interface{}{
		"items": items,
		"pagination": map[string]interface{}{
			"page":        query.Page,
			"page_size":   query.PageSize,
			"total":       total,
			"total_pages": (total + int64(query.PageSize) - 1) / int64(query.PageSize),
		},
	})
}

// LogStats 用量统计
// @Summary 用量统计
// @Tags LLM
// @Produce json
// @Security BearerAuth
// @Param start_time query string false "起始时间"
// @Param end_time query string false "结束时间"
// @Param group_by query string false "分组维度"
// @Success 200 {object} Resp
// @Router /v1/llm/logs/stats [get]
func (h *LLMHandler) LogStats(c *gin.Context) {
	query := repository.LLMCallLogStatsQuery{
		GroupBy: c.DefaultQuery("group_by", "provider"),
	}
	if st := c.Query("start_time"); st != "" {
		if t, err := time.Parse(time.RFC3339, st); err == nil {
			query.StartTime = &t
		}
	}
	if et := c.Query("end_time"); et != "" {
		if t, err := time.Parse(time.RFC3339, et); err == nil {
			query.EndTime = &t
		}
	}
	stats, groups, err := h.svc.LogStats(c.Request.Context(), query)
	if err != nil {
		fail(c, mapLLMErr(err), err.Error())
		return
	}
	ok(c, map[string]interface{}{
		"total_calls":             stats.TotalCalls,
		"success_calls":           stats.SuccessCalls,
		"failed_calls":            stats.FailedCalls,
		"total_tokens":            stats.TotalTokens,
		"total_prompt_tokens":     stats.TotalPromptTokens,
		"total_completion_tokens": stats.TotalCompletionTokens,
		"avg_duration_ms":         stats.AvgDurationMs,
		"groups":                  groups,
	})
}

// ======= 辅助函数 =======

func llmModelToMap(m *model.LLMModel) map[string]interface{} {
	return map[string]interface{}{
		"id":           m.ID,
		"name":         m.Name,
		"provider":     m.Provider,
		"base_url":     m.BaseURL,
		"api_key":      m.MaskedAPIKey(),
		"model":        m.Model,
		"max_tokens":   m.MaxTokens,
		"temperature":  m.Temperature,
		"timeout":      m.Timeout,
		"purpose":      m.Purpose,
		"is_active":    m.IsActive,
		"extra_config": m.ExtraConfig,
		"created_at":   m.CreatedAt,
		"updated_at":   m.UpdatedAt,
	}
}

func mapLLMErr(err error) int {
	if err == nil {
		return 0
	}
	switch err.Error() {
	case "无可用模型配置", "指定的模型配置不存在":
		return 40002
	case "模型配置不存在":
		return 40401
	case "模型限流":
		return 42901
	case "调用超时":
		return 50002
	case "模型调用失败":
		return 50001
	default:
		return 40001
	}
}
