package controllers

import (
	"github.com/zhsyourai/URCF-engine/services/netfilter"
	"github.com/gin-gonic/gin"
)

func NewNetFilterController() *NetFilterController {
	return &NetFilterController{
		service: netfilter.GetInstance(),
	}
}

// NetFilterController is our /netfilter controller.
type NetFilterController struct {
	service netfilter.Service
}

func (c *NetFilterController) Handler(root *gin.RouterGroup) {
}
