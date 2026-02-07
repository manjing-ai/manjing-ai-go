package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"manjing-ai-go/internal/repository"
	"manjing-ai-go/internal/service"
)

// ProjectHandler 项目处理器
type ProjectHandler struct {
	svc service.ProjectService
}

// NewProjectHandler 创建处理器
func NewProjectHandler(svc service.ProjectService) *ProjectHandler {
	return &ProjectHandler{svc: svc}
}

// CreateProjectReq 创建项目请求
type CreateProjectReq struct {
	Name             string `json:"name"`               // 项目名称（必填）
	NarrativeMode    int16  `json:"narrative_mode"`     // 叙事模式（1剧情/2旁白，默认1）
	CoverResourceID  *int64 `json:"cover_resource_id"`  // 封面资源ID（可选）
	VideoAspectRatio string `json:"video_aspect_ratio"` // 视频比例（可选，默认16:9）
	StyleRef         string `json:"style_ref"`          // 风格参考（可选）
}

// UpdateProjectReq 更新项目请求
type UpdateProjectReq struct {
	Name             *string `json:"name"`               // 项目名称（可选）
	NarrativeMode    *int16  `json:"narrative_mode"`     // 叙事模式（可选：1剧情/2旁白）
	CoverResourceID  *int64  `json:"cover_resource_id"`  // 封面资源ID（可选）
	VideoAspectRatio *string `json:"video_aspect_ratio"` // 视频比例（可选）
	StyleRef         *string `json:"style_ref"`          // 风格参考（可选）
	Status           *int16  `json:"status"`             // 状态（可选：1进行中/2归档/3删除）
}

// Create 创建项目
// @Summary 创建项目
// @Tags Project
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body CreateProjectReq true "创建项目"
// @Success 201 {object} Resp
// @Router /v1/projects [post]
func (h *ProjectHandler) Create(c *gin.Context) {
	userID := c.GetInt64("user_id")
	var req CreateProjectReq
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 40001, "参数错误")
		return
	}
	project, err := h.svc.Create(c.Request.Context(), userID, req.Name, req.NarrativeMode, req.CoverResourceID, req.VideoAspectRatio, req.StyleRef)
	if err != nil {
		fail(c, mapProjectErr(err), err.Error())
		return
	}
	c.JSON(http.StatusCreated, Resp{
		Code:    0,
		Message: "success",
		Data: map[string]interface{}{
			"id":                 project.ID,
			"name":               project.Name,
			"narrative_mode":     project.NarrativeMode,
			"cover_resource_id":  project.CoverResourceID,
			"video_aspect_ratio": project.VideoAspectRatio,
			"style_ref":          project.StyleRef,
			"status":             project.Status,
			"created_at":         project.CreatedAt,
			"updated_at":         project.UpdatedAt,
		},
	})
}

// List 获取项目列表
// @Summary 获取项目列表
// @Tags Project
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码"
// @Param page_size query int false "每页数量"
// @Param status query int false "状态"
// @Param keyword query string false "关键词"
// @Param sort query string false "排序"
// @Success 200 {object} Resp
// @Router /v1/projects [get]
func (h *ProjectHandler) List(c *gin.Context) {
	userID := c.GetInt64("user_id")
	query := repository.ProjectListQuery{
		Page:     parseIntDef(c.Query("page"), 1),
		PageSize: parseIntDef(c.Query("page_size"), 20),
		Status:   parseIntDef(c.Query("status"), 0),
		Keyword:  c.Query("keyword"),
		Sort:     c.Query("sort"),
	}
	items, total, err := h.svc.List(c.Request.Context(), userID, query)
	if err != nil {
		fail(c, mapProjectErr(err), err.Error())
		return
	}
	list := make([]map[string]interface{}, 0, len(items))
	for _, it := range items {
		list = append(list, map[string]interface{}{
			"id":                 it.ID,
			"name":               it.Name,
			"status":             it.Status,
			"narrative_mode":     it.NarrativeMode,
			"cover_resource_id":  it.CoverResourceID,
			"video_aspect_ratio": it.VideoAspectRatio,
			"style_ref":          it.StyleRef,
			"updated_at":         it.UpdatedAt,
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

// Detail 获取项目详情
// @Summary 获取项目详情
// @Tags Project
// @Produce json
// @Security BearerAuth
// @Param id path int true "项目ID"
// @Success 200 {object} Resp
// @Router /v1/projects/{id} [get]
func (h *ProjectHandler) Detail(c *gin.Context) {
	userID := c.GetInt64("user_id")
	id, err := parseID(c.Param("id"))
	if err != nil {
		fail(c, 40001, "参数错误")
		return
	}
	project, err := h.svc.Get(c.Request.Context(), userID, id)
	if err != nil {
		fail(c, mapProjectErr(err), err.Error())
		return
	}
	ok(c, map[string]interface{}{
		"id":                 project.ID,
		"name":               project.Name,
		"narrative_mode":     project.NarrativeMode,
		"cover_resource_id":  project.CoverResourceID,
		"video_aspect_ratio": project.VideoAspectRatio,
		"style_ref":          project.StyleRef,
		"status":             project.Status,
		"created_at":         project.CreatedAt,
		"updated_at":         project.UpdatedAt,
	})
}

// Update 更新项目
// @Summary 更新项目
// @Tags Project
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "项目ID"
// @Param body body UpdateProjectReq true "更新项目"
// @Success 200 {object} Resp
// @Router /v1/projects/{id} [put]
func (h *ProjectHandler) Update(c *gin.Context) {
	userID := c.GetInt64("user_id")
	id, err := parseID(c.Param("id"))
	if err != nil {
		fail(c, 40001, "参数错误")
		return
	}
	var req UpdateProjectReq
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 40001, "参数错误")
		return
	}
	project, err := h.svc.Update(c.Request.Context(), userID, id, service.ProjectUpdate{
		Name:             req.Name,
		NarrativeMode:    req.NarrativeMode,
		CoverResourceID:  req.CoverResourceID,
		VideoAspectRatio: req.VideoAspectRatio,
		StyleRef:         req.StyleRef,
		Status:           req.Status,
	})
	if err != nil {
		fail(c, mapProjectErr(err), err.Error())
		return
	}
	ok(c, map[string]interface{}{
		"id":                 project.ID,
		"name":               project.Name,
		"narrative_mode":     project.NarrativeMode,
		"cover_resource_id":  project.CoverResourceID,
		"video_aspect_ratio": project.VideoAspectRatio,
		"style_ref":          project.StyleRef,
		"status":             project.Status,
		"updated_at":         project.UpdatedAt,
	})
}

// Delete 删除项目
// @Summary 删除项目
// @Tags Project
// @Produce json
// @Security BearerAuth
// @Param id path int true "项目ID"
// @Success 204 {object} Resp
// @Router /v1/projects/{id} [delete]
func (h *ProjectHandler) Delete(c *gin.Context) {
	userID := c.GetInt64("user_id")
	id, err := parseID(c.Param("id"))
	if err != nil {
		fail(c, 40001, "参数错误")
		return
	}
	if err := h.svc.Delete(c.Request.Context(), userID, id); err != nil {
		fail(c, mapProjectErr(err), err.Error())
		return
	}
	c.Status(http.StatusNoContent)
}

// Restore 恢复项目
// @Summary 恢复项目
// @Tags Project
// @Produce json
// @Security BearerAuth
// @Param id path int true "项目ID"
// @Success 200 {object} Resp
// @Router /v1/projects/{id}/restore [post]
func (h *ProjectHandler) Restore(c *gin.Context) {
	userID := c.GetInt64("user_id")
	id, err := parseID(c.Param("id"))
	if err != nil {
		fail(c, 40001, "参数错误")
		return
	}
	project, err := h.svc.Restore(c.Request.Context(), userID, id)
	if err != nil {
		fail(c, mapProjectErr(err), err.Error())
		return
	}
	ok(c, map[string]interface{}{
		"id":         project.ID,
		"status":     project.Status,
		"updated_at": project.UpdatedAt,
	})
}

// Archive 归档项目
// @Summary 归档项目
// @Tags Project
// @Produce json
// @Security BearerAuth
// @Param id path int true "项目ID"
// @Success 200 {object} Resp
// @Router /v1/projects/{id}/archive [post]
func (h *ProjectHandler) Archive(c *gin.Context) {
	userID := c.GetInt64("user_id")
	id, err := parseID(c.Param("id"))
	if err != nil {
		fail(c, 40001, "参数错误")
		return
	}
	project, err := h.svc.Archive(c.Request.Context(), userID, id)
	if err != nil {
		fail(c, mapProjectErr(err), err.Error())
		return
	}
	ok(c, map[string]interface{}{
		"id":         project.ID,
		"status":     project.Status,
		"updated_at": project.UpdatedAt,
	})
}

func parseIntDef(s string, def int) int {
	if s == "" {
		return def
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return v
}

func mapProjectErr(err error) int {
	if err == nil {
		return 0
	}
	switch err.Error() {
	case "未授权", "无权访问":
		return 40301
	case "项目不存在":
		return 40401
	default:
		return 40001
	}
}
