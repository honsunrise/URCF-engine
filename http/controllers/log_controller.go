package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/zhsyourai/URCF-engine/http/controllers/shard"
	"github.com/zhsyourai/URCF-engine/http/gin-jwt"
	"github.com/zhsyourai/URCF-engine/services/log"
	"net/http"
	"strconv"
)

func NewLogController(middleware *gin_jwt.JwtMiddleware) *LogController {
	return &LogController{
		service:    log.GetInstance(),
		middleware: middleware,
	}
}

// LogController is our /log controller.
type LogController struct {
	service    log.Service
	middleware *gin_jwt.JwtMiddleware
}

func (c *LogController) Handler(root *gin.RouterGroup) {
	root.GET("/list", c.ListLogHandler)
	root.DELETE("/clean/*id", c.CleanLogHandler)
}

func (c *LogController) ListLogHandler(ctx *gin.Context) {
	var paging shard.Paging
	if ctx.BindQuery(&paging) != nil {
		return
	}

	total, logs, err := c.service.ListAll(paging.Page, paging.Size, paging.Sort, paging.Order)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, &shard.LogWithCount{
		TotalCount: total,
		Items:      logs,
	})
}

func (c *LogController) CleanLogHandler(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 0)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.service.Clean(id)
	ctx.Status(http.StatusOK)
}
