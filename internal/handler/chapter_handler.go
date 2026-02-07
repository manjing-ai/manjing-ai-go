package handler

import (
	"net/http"

	"manjing-ai-go/internal/repository"
	"manjing-ai-go/internal/service"

	"github.com/gin-gonic/gin"
)

// ChapterHandler 章节处理器
type ChapterHandler struct {
	svc service.ChapterService
}

// NewChapterHandler 创建处理器
func NewChapterHandler(svc service.ChapterService) *ChapterHandler {
	return &ChapterHandler{svc: svc}
}

// CreateChapterReq 创建章节请求
type CreateChapterReq struct {
	ProjectID  int64  `json:"project_id"`  // 项目ID（必填）
	Name       string `json:"name"`        // 章节名称（必填）
	Content    string `json:"content"`     // 章节内容（可选）
	Summary    string `json:"summary"`     // 章节摘要（可选）
	OrderIndex int    `json:"order_index"` // 章节顺序（可选）
}

// UpdateChapterReq 更新章节请求
type UpdateChapterReq struct {
	Name       *string `json:"name"`        // 章节名称（可选）
	Content    *string `json:"content"`     // 章节内容（可选）
	Summary    *string `json:"summary"`     // 章节摘要（可选）
	OrderIndex *int    `json:"order_index"` // 章节顺序（可选）
	Status     *int16  `json:"status"`      // 状态（可选：1进行中/2归档/3删除）
}

// Create 创建章节
// @Summary 创建章节
// @Tags Chapter
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body CreateChapterReq true "创建章节"
// @Success 201 {object} Resp
// @Router /v1/chapters [post]
func (h *ChapterHandler) Create(c *gin.Context) {
	userID := c.GetInt64("user_id")
	var req CreateChapterReq
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 40001, "参数错误")
		return
	}
	chapter, err := h.svc.Create(c.Request.Context(), userID, req.ProjectID, req.Name, req.Content, req.Summary, req.OrderIndex)
	if err != nil {
		fail(c, mapChapterErr(err), err.Error())
		return
	}
	c.JSON(http.StatusCreated, Resp{
		Code:    0,
		Message: "success",
		Data: map[string]interface{}{
			"id":          chapter.ID,
			"project_id":  chapter.ProjectID,
			"name":        chapter.Name,
			"content":     chapter.Content,
			"summary":     chapter.Summary,
			"order_index": chapter.OrderIndex,
			"status":      chapter.Status,
			"created_at":  chapter.CreatedAt,
			"updated_at":  chapter.UpdatedAt,
		},
	})
}

// List 获取章节列表
// @Summary 获取章节列表
// @Tags Chapter
// @Produce json
// @Security BearerAuth
// @Param project_id query int true "项目ID"
// @Param page query int false "页码"
// @Param page_size query int false "每页数量"
// @Param status query int false "状态"
// @Param keyword query string false "关键词"
// @Param sort query string false "排序"
// @Success 200 {object} Resp
// @Router /v1/chapters [get]
func (h *ChapterHandler) List(c *gin.Context) {
	userID := c.GetInt64("user_id")
	projectID, err := parseID(c.Query("project_id"))
	if err != nil || projectID == 0 {
		fail(c, 40001, "参数错误")
		return
	}
	query := repository.ChapterListQuery{
		Page:     parseIntDef(c.Query("page"), 1),
		PageSize: parseIntDef(c.Query("page_size"), 20),
		Status:   parseIntDef(c.Query("status"), 0),
		Keyword:  c.Query("keyword"),
		Sort:     c.Query("sort"),
	}
	items, total, err := h.svc.List(c.Request.Context(), userID, projectID, query)
	if err != nil {
		fail(c, mapChapterErr(err), err.Error())
		return
	}
	list := make([]map[string]interface{}, 0, len(items))
	for _, it := range items {
		list = append(list, map[string]interface{}{
			"id":          it.ID,
			"project_id":  it.ProjectID,
			"name":        it.Name,
			"summary":     it.Summary,
			"order_index": it.OrderIndex,
			"status":      it.Status,
			"updated_at":  it.UpdatedAt,
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

// Detail 获取章节详情
// @Summary 获取章节详情
// @Tags Chapter
// @Produce json
// @Security BearerAuth
// @Param id path int true "章节ID"
// @Success 200 {object} Resp
// @Router /v1/chapters/{id} [get]
func (h *ChapterHandler) Detail(c *gin.Context) {
	userID := c.GetInt64("user_id")
	id, err := parseID(c.Param("id"))
	if err != nil {
		fail(c, 40001, "参数错误")
		return
	}
	chapter, err := h.svc.Get(c.Request.Context(), userID, id)
	if err != nil {
		fail(c, mapChapterErr(err), err.Error())
		return
	}
	ok(c, map[string]interface{}{
		"id":          chapter.ID,
		"project_id":  chapter.ProjectID,
		"name":        chapter.Name,
		"content":     chapter.Content,
		"summary":     chapter.Summary,
		"order_index": chapter.OrderIndex,
		"status":      chapter.Status,
		"created_at":  chapter.CreatedAt,
		"updated_at":  chapter.UpdatedAt,
	})
}

// Update 更新章节
// @Summary 更新章节
// @Tags Chapter
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "章节ID"
// @Param body body UpdateChapterReq true "更新章节"
// @Success 200 {object} Resp
// @Router /v1/chapters/{id} [put]
func (h *ChapterHandler) Update(c *gin.Context) {
	userID := c.GetInt64("user_id")
	id, err := parseID(c.Param("id"))
	if err != nil {
		fail(c, 40001, "参数错误")
		return
	}
	var req UpdateChapterReq
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 40001, "参数错误")
		return
	}
	chapter, err := h.svc.Update(c.Request.Context(), userID, id, service.ChapterUpdate{
		Name:       req.Name,
		Content:    req.Content,
		Summary:    req.Summary,
		OrderIndex: req.OrderIndex,
		Status:     req.Status,
	})
	if err != nil {
		fail(c, mapChapterErr(err), err.Error())
		return
	}
	ok(c, map[string]interface{}{
		"id":          chapter.ID,
		"name":        chapter.Name,
		"order_index": chapter.OrderIndex,
		"status":      chapter.Status,
		"updated_at":  chapter.UpdatedAt,
	})
}

// Delete 删除章节
// @Summary 删除章节
// @Tags Chapter
// @Produce json
// @Security BearerAuth
// @Param id path int true "章节ID"
// @Success 204 {object} Resp
// @Router /v1/chapters/{id} [delete]
func (h *ChapterHandler) Delete(c *gin.Context) {
	userID := c.GetInt64("user_id")
	id, err := parseID(c.Param("id"))
	if err != nil {
		fail(c, 40001, "参数错误")
		return
	}
	if err := h.svc.Delete(c.Request.Context(), userID, id); err != nil {
		fail(c, mapChapterErr(err), err.Error())
		return
	}
	c.Status(http.StatusNoContent)
}

// Restore 恢复章节
// @Summary 恢复章节
// @Tags Chapter
// @Produce json
// @Security BearerAuth
// @Param id path int true "章节ID"
// @Success 200 {object} Resp
// @Router /v1/chapters/{id}/restore [post]
func (h *ChapterHandler) Restore(c *gin.Context) {
	userID := c.GetInt64("user_id")
	id, err := parseID(c.Param("id"))
	if err != nil {
		fail(c, 40001, "参数错误")
		return
	}
	chapter, err := h.svc.Restore(c.Request.Context(), userID, id)
	if err != nil {
		fail(c, mapChapterErr(err), err.Error())
		return
	}
	ok(c, map[string]interface{}{
		"id":         chapter.ID,
		"status":     chapter.Status,
		"updated_at": chapter.UpdatedAt,
	})
}

// Archive 归档章节
// @Summary 归档章节
// @Tags Chapter
// @Produce json
// @Security BearerAuth
// @Param id path int true "章节ID"
// @Success 200 {object} Resp
// @Router /v1/chapters/{id}/archive [post]
func (h *ChapterHandler) Archive(c *gin.Context) {
	userID := c.GetInt64("user_id")
	id, err := parseID(c.Param("id"))
	if err != nil {
		fail(c, 40001, "参数错误")
		return
	}
	chapter, err := h.svc.Archive(c.Request.Context(), userID, id)
	if err != nil {
		fail(c, mapChapterErr(err), err.Error())
		return
	}
	ok(c, map[string]interface{}{
		"id":         chapter.ID,
		"status":     chapter.Status,
		"updated_at": chapter.UpdatedAt,
	})
}

func mapChapterErr(err error) int {
	if err == nil {
		return 0
	}
	switch err.Error() {
	case "未授权", "无权访问":
		return 40301
	case "项目不存在":
		return 40402
	case "章节不存在":
		return 40401
	default:
		return 40001
	}
}
