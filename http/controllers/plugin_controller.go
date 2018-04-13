package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/zhsyourai/URCF-engine/http/gin-jwt"
	"github.com/zhsyourai/URCF-engine/services/plugin"
)

func NewPluginController(middleware *gin_jwt.JwtMiddleware) *PluginController {
	return &PluginController{
		service:    plugin.GetInstance(),
		middleware: middleware,
	}
}

// PluginController is our /plugin controller.
type PluginController struct {
	service    plugin.Service
	middleware *gin_jwt.JwtMiddleware
}

func (c *PluginController) Handler(root *gin.RouterGroup) {
}
