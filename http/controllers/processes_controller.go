package controllers

import (
	"github.com/zhsyourai/URCF-engine/services/processes"
	"github.com/gin-gonic/gin"
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

