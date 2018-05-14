package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/zhsyourai/URCF-engine/http/gin-jwt"
	"github.com/zhsyourai/URCF-engine/services/plugin"
	"github.com/zhsyourai/URCF-engine/utils"
	"net/http"
	"path"
	"sync"
)

func NewWebsController(middleware *gin_jwt.JwtMiddleware) *WebsController {
	return &WebsController{
		service:    plugin.GetInstance(),
		middleware: middleware,
	}
}

// WebsController is our /plugin controller.
type WebsController struct {
	service       plugin.Service
	middleware    *gin_jwt.JwtMiddleware
	fileServerMap sync.Map
}

func (c *WebsController) Handler(root *gin.RouterGroup) {
	root.GET("/:plugin/*other", c.PluginRootHandler(root.BasePath()))
	root.HEAD("/:plugin/*other", c.PluginRootHandler(root.BasePath()))
}

func (c *WebsController) PluginRootHandler(basePath string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		pluginStr := ctx.Param("plugin")

		p, err := c.service.GetByName(pluginStr)
		if err != nil {
			ctx.AbortWithError(http.StatusNotFound, err)
			return
		}

		absolutePath := utils.JoinPath(basePath, pluginStr)
		tmp, ok := c.fileServerMap.Load(pluginStr)
		if ok {
			handler := tmp.(http.Handler)
			ctx.Writer.WriteHeader(404)
			handler.ServeHTTP(ctx.Writer, ctx.Request)
		} else {
			targetDir := path.Join(p.InstallDir, p.WebsDir)
			fileServer := http.StripPrefix(absolutePath, http.FileServer(gin.Dir(targetDir, false)))
			ctx.Writer.WriteHeader(404)
			fileServer.ServeHTTP(ctx.Writer, ctx.Request)
			c.fileServerMap.Store(pluginStr, fileServer)
		}
	}
}
