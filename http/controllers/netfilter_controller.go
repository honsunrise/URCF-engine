package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/zhsyourai/URCF-engine/services/netfilter"
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
