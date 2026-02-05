package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"manjing-ai-go/internal/repository"
	"manjing-ai-go/internal/service"
)

// ResourceHandler 资源处理器
type ResourceHandler struct {
	svc service.ResourceService
}

// NewResourceHandler 创建处理器
func NewResourceHandler(svc service.ResourceService) *ResourceHandler {
	return &ResourceHandler{svc: svc}
}

// UploadReq 上传请求
type UploadReq struct {
	Name      string `form:"name"`       // 资源名称（可选）
	Type      string `form:"type"`       // 资源类型（可选）
	Category  string `form:"category"`   // 分类（可选）
	ExtraData string `form:"extra_data"` // 扩展数据（JSON字符串）
}

// UpdateReq 更新请求
type UpdateReq struct {
	Name      string                 `json:"name"`       // 资源名称（可选）
	Category  string                 `json:"category"`   // 分类（可选）
	ExtraData map[string]interface{} `json:"extra_data"` // 扩展数据（JSON）
}

// Upload 上传资源
// @Summary 上传资源
// @Tags Resource
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param file formData file true "资源文件"
// @Param name formData string false "资源名称"
// @Param type formData string false "资源类型"
// @Param category formData string false "分类"
// @Param extra_data formData string false "扩展数据(JSON)"
// @Success 201 {object} Resp
// @Router /v1/resources [post]
func (h *ResourceHandler) Upload(c *gin.Context) {
	userID := c.GetInt64("user_id")
	file, err := c.FormFile("file")
	if err != nil {
		fail(c, 40001, "文件不能为空")
		return
	}

	f, err := file.Open()
	if err != nil {
		fail(c, 50001, "文件读取失败")
		return
	}
	defer f.Close()

	buf, err := io.ReadAll(f)
	if err != nil {
		fail(c, 50001, "文件读取失败")
		return
	}

	var req UploadReq
	_ = c.ShouldBind(&req)

	res, url, err := h.svc.Upload(c.Request.Context(), userID, file.Filename, buf, req.Name, req.Type, req.Category, req.ExtraData)
	if err != nil {
		fail(c, 40001, err.Error())
		return
	}

	c.JSON(http.StatusCreated, Resp{
		Code:    0,
		Message: "success",
		Data: map[string]interface{}{
			"id":         res.ID,
			"name":       res.Name,
			"type":       res.Type,
			"category":   res.Category,
			"url":        url,
			"file_name":  res.FileName,
			"file_ext":   res.FileExt,
			"mime_type":  res.MimeType,
			"width":      res.Width,
			"height":     res.Height,
			"aspect":     res.Aspect,
			"size_bytes": res.SizeBytes,
			"status":     res.Status,
			"extra_data": res.ExtraData,
			"created_at": res.CreatedAt,
		},
	})
}

// List 获取资源列表
// @Summary 获取资源列表
// @Tags Resource
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码"
// @Param page_size query int false "每页数量"
// @Param type query string false "类型"
// @Param category query string false "分类"
// @Param status query string false "状态"
// @Param keyword query string false "关键词"
// @Param sort query string false "排序"
// @Success 200 {object} Resp
// @Router /v1/resources [get]
func (h *ResourceHandler) List(c *gin.Context) {
	userID := c.GetInt64("user_id")
	query := repository.ResourceListQuery{
		Page:     parseInt(c.Query("page"), 1),
		PageSize: parseInt(c.Query("page_size"), 20),
		Type:     c.Query("type"),
		Category: c.Query("category"),
		Status:   c.Query("status"),
		Keyword:  c.Query("keyword"),
		Sort:     c.Query("sort"),
	}

	items, total, err := h.svc.List(c.Request.Context(), userID, query)
	if err != nil {
		fail(c, 50001, err.Error())
		return
	}

	list := make([]map[string]interface{}, 0, len(items))
	for _, it := range items {
		list = append(list, map[string]interface{}{
			"id":         it.Resource.ID,
			"name":       it.Resource.Name,
			"type":       it.Resource.Type,
			"category":   it.Resource.Category,
			"url":        it.URL,
			"size_bytes": it.Resource.SizeBytes,
			"created_at": it.Resource.CreatedAt,
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

// Detail 获取资源详情
// @Summary 获取资源详情
// @Tags Resource
// @Produce json
// @Security BearerAuth
// @Param id path int true "资源ID"
// @Success 200 {object} Resp
// @Router /v1/resources/{id} [get]
func (h *ResourceHandler) Detail(c *gin.Context) {
	userID := c.GetInt64("user_id")
	id, err := parseID(c.Param("id"))
	if err != nil {
		fail(c, 40001, "参数错误")
		return
	}
	res, url, err := h.svc.Get(c.Request.Context(), userID, id)
	if err != nil {
		fail(c, 40401, err.Error())
		return
	}

	ok(c, map[string]interface{}{
		"id":         res.ID,
		"name":       res.Name,
		"type":       res.Type,
		"category":   res.Category,
		"url":        url,
		"file_name":  res.FileName,
		"file_ext":   res.FileExt,
		"mime_type":  res.MimeType,
		"width":      res.Width,
		"height":     res.Height,
		"aspect":     res.Aspect,
		"size_bytes": res.SizeBytes,
		"status":     res.Status,
		"extra_data": res.ExtraData,
		"created_at": res.CreatedAt,
		"updated_at": res.UpdatedAt,
	})
}

// Update 更新资源信息
// @Summary 更新资源信息
// @Tags Resource
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "资源ID"
// @Param body body UpdateReq true "更新信息"
// @Success 200 {object} Resp
// @Router /v1/resources/{id} [put]
func (h *ResourceHandler) Update(c *gin.Context) {
	userID := c.GetInt64("user_id")
	id, err := parseID(c.Param("id"))
	if err != nil {
		fail(c, 40001, "参数错误")
		return
	}
	var req UpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 40001, "参数错误")
		return
	}
	extra := ""
	if req.ExtraData != nil {
		if b, err := json.Marshal(req.ExtraData); err == nil {
			extra = string(b)
		}
	}
	res, err := h.svc.Update(c.Request.Context(), userID, id, req.Name, req.Category, extra)
	if err != nil {
		fail(c, 40001, err.Error())
		return
	}
	ok(c, map[string]interface{}{
		"id":         res.ID,
		"name":       res.Name,
		"category":   res.Category,
		"extra_data": res.ExtraData,
		"updated_at": res.UpdatedAt,
	})
}

// Delete 删除资源
// @Summary 删除资源
// @Tags Resource
// @Produce json
// @Security BearerAuth
// @Param id path int true "资源ID"
// @Param hard query bool false "是否硬删除"
// @Success 204 {object} Resp
// @Router /v1/resources/{id} [delete]
func (h *ResourceHandler) Delete(c *gin.Context) {
	userID := c.GetInt64("user_id")
	id, err := parseID(c.Param("id"))
	if err != nil {
		fail(c, 40001, "参数错误")
		return
	}
	hard := c.DefaultQuery("hard", "false") == "true"
	if err := h.svc.Delete(c.Request.Context(), userID, id, hard); err != nil {
		fail(c, 40001, err.Error())
		return
	}
	c.Status(http.StatusNoContent)
}

func parseInt(s string, def int) int {
	if s == "" {
		return def
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return v
}
