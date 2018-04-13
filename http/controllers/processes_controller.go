package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/zhsyourai/URCF-engine/services/processes"
)

func NewProcessesController() *ProcessesController {
	return &ProcessesController{
		service: processes.GetInstance(),
	}
}

// ProcessesController is our /processes controller.
type ProcessesController struct {
	service processes.Service
}

func (c *ProcessesController) Handler(root *gin.RouterGroup) {
}
