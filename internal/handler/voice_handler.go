package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"manjing-ai-go/internal/repository"
	"manjing-ai-go/internal/service"
)

// VoiceHandler 声音处理器
type VoiceHandler struct {
	svc service.VoiceService
}

// NewVoiceHandler 创建声音处理器
func NewVoiceHandler(svc service.VoiceService) *VoiceHandler {
	return &VoiceHandler{svc: svc}
}

// CreateVoiceReq 创建声音请求
type CreateVoiceReq struct {
	Name      string `json:"name"`       // 声音名称（必填，1-64字符）
	AgeGroup  int16  `json:"age_group"`  // 年龄段（必填：1儿童/2少年/3青年/4中年/5老年）
	Gender    int16  `json:"gender"`     // 性别（必填：1男/2女）
	Dialect   int16  `json:"dialect"`    // 方言口音（必填：1标准普通话/2东北话/3四川话/4粤语/5台湾腔/6港式普通话/7外国口音）
	Tone      int16  `json:"tone"`       // 音色（必填：1标准/2清亮/3浑厚/4沙哑/5柔和/6尖细/7气声/8鼻音/9金属）
	SampleURL string `json:"sample_url"` // 试听音频URL（可选）
	Type      int16  `json:"type"`       // 类型（必填：1官方/2用户）
}

// UpdateVoiceReq 更新声音请求
type UpdateVoiceReq struct {
	Name      *string `json:"name"`       // 声音名称（可选）
	AgeGroup  *int16  `json:"age_group"`  // 年龄段（可选）
	Gender    *int16  `json:"gender"`     // 性别（可选）
	Dialect   *int16  `json:"dialect"`    // 方言口音（可选）
	Tone      *int16  `json:"tone"`       // 音色（可选）
	SampleURL *string `json:"sample_url"` // 试听音频URL（可选）
}

// Create 创建声音
// @Summary 创建声音
// @Tags Voice
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body CreateVoiceReq true "创建声音"
// @Success 201 {object} Resp
// @Router /v1/voices [post]
func (h *VoiceHandler) Create(c *gin.Context) {
	userID := c.GetInt64("user_id")
	var req CreateVoiceReq
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 40001, "参数错误")
		return
	}
	voice, err := h.svc.Create(c.Request.Context(), userID, service.VoiceCreate{
		Name:      req.Name,
		AgeGroup:  req.AgeGroup,
		Gender:    req.Gender,
		Dialect:   req.Dialect,
		Tone:      req.Tone,
		SampleURL: req.SampleURL,
		Type:      req.Type,
	})
	if err != nil {
		fail(c, mapVoiceErr(err), err.Error())
		return
	}
	c.JSON(http.StatusCreated, Resp{
		Code:    0,
		Message: "success",
		Data: map[string]interface{}{
			"id":            voice.ID,
			"name":          voice.Name,
			"age_group":     voice.AgeGroup,
			"gender":        voice.Gender,
			"dialect":       voice.Dialect,
			"tone":          voice.Tone,
			"sample_url":    voice.SampleURL,
			"type":          voice.Type,
			"owner_user_id": voice.OwnerUserID,
			"created_at":    voice.CreatedAt,
			"updated_at":    voice.UpdatedAt,
		},
	})
}

// List 获取声音列表
// @Summary 获取声音列表
// @Tags Voice
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码"
// @Param page_size query int false "每页数量"
// @Param type query int false "类型：1官方/2用户"
// @Param age_group query int false "年龄段"
// @Param gender query int false "性别"
// @Param dialect query int false "方言口音"
// @Param tone query int false "音色"
// @Param keyword query string false "名称关键词"
// @Param sort query string false "排序"
// @Success 200 {object} Resp
// @Router /v1/voices [get]
func (h *VoiceHandler) List(c *gin.Context) {
	userID := c.GetInt64("user_id")
	query := repository.VoiceListQuery{
		Page:     parseIntDef(c.Query("page"), 1),
		PageSize: parseIntDef(c.Query("page_size"), 20),
		Type:     parseIntDef(c.Query("type"), 0),
		AgeGroup: parseIntDef(c.Query("age_group"), 0),
		Gender:   parseIntDef(c.Query("gender"), 0),
		Dialect:  parseIntDef(c.Query("dialect"), 0),
		Tone:     parseIntDef(c.Query("tone"), 0),
		Keyword:  c.Query("keyword"),
		Sort:     c.Query("sort"),
	}
	items, total, err := h.svc.List(c.Request.Context(), userID, query)
	if err != nil {
		fail(c, mapVoiceErr(err), err.Error())
		return
	}
	list := make([]map[string]interface{}, 0, len(items))
	for _, it := range items {
		list = append(list, map[string]interface{}{
			"id":         it.ID,
			"name":       it.Name,
			"age_group":  it.AgeGroup,
			"gender":     it.Gender,
			"dialect":    it.Dialect,
			"tone":       it.Tone,
			"sample_url": it.SampleURL,
			"type":       it.Type,
			"created_at": it.CreatedAt,
		})
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

// Detail 获取声音详情
// @Summary 获取声音详情
// @Tags Voice
// @Produce json
// @Security BearerAuth
// @Param id path int true "声音ID"
// @Success 200 {object} Resp
// @Router /v1/voices/{id} [get]
func (h *VoiceHandler) Detail(c *gin.Context) {
	userID := c.GetInt64("user_id")
	id, err := parseID(c.Param("id"))
	if err != nil {
		fail(c, 40001, "参数错误")
		return
	}
	voice, err := h.svc.Get(c.Request.Context(), userID, id)
	if err != nil {
		fail(c, mapVoiceErr(err), err.Error())
		return
	}
	ok(c, map[string]interface{}{
		"id":            voice.ID,
		"name":          voice.Name,
		"age_group":     voice.AgeGroup,
		"gender":        voice.Gender,
		"dialect":       voice.Dialect,
		"tone":          voice.Tone,
		"sample_url":    voice.SampleURL,
		"type":          voice.Type,
		"owner_user_id": voice.OwnerUserID,
		"created_at":    voice.CreatedAt,
		"updated_at":    voice.UpdatedAt,
	})
}

// Update 更新声音
// @Summary 更新声音
// @Tags Voice
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "声音ID"
// @Param body body UpdateVoiceReq true "更新声音"
// @Success 200 {object} Resp
// @Router /v1/voices/{id} [put]
func (h *VoiceHandler) Update(c *gin.Context) {
	userID := c.GetInt64("user_id")
	id, err := parseID(c.Param("id"))
	if err != nil {
		fail(c, 40001, "参数错误")
		return
	}
	var req UpdateVoiceReq
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 40001, "参数错误")
		return
	}
	voice, err := h.svc.Update(c.Request.Context(), userID, id, service.VoiceUpdate{
		Name:      req.Name,
		AgeGroup:  req.AgeGroup,
		Gender:    req.Gender,
		Dialect:   req.Dialect,
		Tone:      req.Tone,
		SampleURL: req.SampleURL,
	})
	if err != nil {
		fail(c, mapVoiceErr(err), err.Error())
		return
	}
	ok(c, map[string]interface{}{
		"id":            voice.ID,
		"name":          voice.Name,
		"age_group":     voice.AgeGroup,
		"gender":        voice.Gender,
		"dialect":       voice.Dialect,
		"tone":          voice.Tone,
		"sample_url":    voice.SampleURL,
		"type":          voice.Type,
		"owner_user_id": voice.OwnerUserID,
		"updated_at":    voice.UpdatedAt,
	})
}

// Delete 删除声音
// @Summary 删除声音
// @Tags Voice
// @Produce json
// @Security BearerAuth
// @Param id path int true "声音ID"
// @Success 204 {object} Resp
// @Router /v1/voices/{id} [delete]
func (h *VoiceHandler) Delete(c *gin.Context) {
	userID := c.GetInt64("user_id")
	id, err := parseID(c.Param("id"))
	if err != nil {
		fail(c, 40001, "参数错误")
		return
	}
	if err := h.svc.Delete(c.Request.Context(), userID, id); err != nil {
		fail(c, mapVoiceErr(err), err.Error())
		return
	}
	c.Status(http.StatusNoContent)
}

func mapVoiceErr(err error) int {
	if err == nil {
		return 0
	}
	switch err.Error() {
	case "未授权", "无权访问":
		return 40301
	case "声音不存在":
		return 40401
	default:
		return 40001
	}
}
