package controllers

import (
	"github.com/zhsyourai/URCF-engine/services/log"
	"github.com/gin-gonic/gin"
	"github.com/zhsyourai/URCF-engine/http/gin-jwt"
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

}


func (c *LogController) CleanLogHandler(ctx *gin.Context) {

}