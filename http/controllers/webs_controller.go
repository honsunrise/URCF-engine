package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/zhsyourai/URCF-engine/http/gin-jwt"
	"github.com/zhsyourai/URCF-engine/services/plugin"
	"net/http"
	"path"
)

func NewWebsController(middleware *gin_jwt.JwtMiddleware) *WebsController {
	return &WebsController{
		service:    plugin.GetInstance(),
		middleware: middleware,
	}
}

// WebsController is our /plugin controller.
type WebsController struct {
	service    plugin.Service
	middleware *gin_jwt.JwtMiddleware
}

func (c *WebsController) Handler(root *gin.RouterGroup) {
	root.GET("/:plugin/*other", c.PluginRootHandler)
}

func (c *WebsController) PluginRootHandler(ctx *gin.Context) {
	pluginStr := ctx.Param("plugin")
	otherStr := ctx.Param("other")

	p, err := c.service.GetByName(pluginStr)
	if err != nil {
		ctx.AbortWithError(http.StatusNotFound, err)
		return
	}
	resultFile := path.Join(p.InstallDir, p.WebsDir, otherStr)
	ctx.File(resultFile)
}
