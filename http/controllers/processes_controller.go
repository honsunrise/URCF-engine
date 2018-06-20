package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/zhsyourai/URCF-engine/services/processes"
	"net/http"
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
	root.GET("/list", c.ListHandler)
	root.DELETE("/*name", c.KillHandler)
}

func (c *ProcessesController) ListHandler(ctx *gin.Context) {
	result := c.service.ListAll()
	ctx.JSON(http.StatusOK, &result)
}

func (c *ProcessesController) KillHandler(ctx *gin.Context) {
	name := ctx.Param("name")
	c.service.Kill(name)
	ctx.Status(http.StatusOK)
}
