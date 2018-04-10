package controllers

import (
	"github.com/zhsyourai/URCF-engine/services/log"
	"github.com/gin-gonic/gin"
)

func NewLogController() *LogController {
	return &LogController{
		service: log.GetInstance(),
	}
}

// LogController is our /log controller.
type LogController struct {
	service log.Service
}

func (c *LogController) Handler(root *gin.RouterGroup) {
}
