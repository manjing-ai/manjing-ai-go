package handler

import "github.com/gin-gonic/gin"

// Resp 统一响应结构
type Resp struct {
	Code    int         `json:"code"`    // 错误码，0表示成功
	Message string      `json:"message"` // 状态描述
	Data    interface{} `json:"data"`    // 业务数据
}

func ok(c *gin.Context, data interface{}) {
	c.JSON(200, Resp{
		Code:    0,
		Message: "ok",
		Data:    data,
	})
}

func fail(c *gin.Context, code int, msg string) {
	c.JSON(200, Resp{
		Code:    code,
		Message: msg,
		Data:    map[string]interface{}{},
	})
}
