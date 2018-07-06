package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/zhsyourai/URCF-engine/http/controllers/shard"
	"github.com/zhsyourai/URCF-engine/http/gin-jwt"
	"github.com/zhsyourai/URCF-engine/services/plugin"
	"net/http"
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
	root.GET("", c.ListPluginHandler)
	root.GET("/:name", c.GetPluginHandler)
	root.GET("/:name/commands", c.GetPluginCommandsHandler)
	root.GET("/:name/start", c.StartPluginHandler)
	root.GET("/:name/stop", c.StopPluginHandler)
	root.POST("/:name/:command", c.ExecPluginCommandHandler)
	root.POST("", c.InstallPluginHandler)
	root.DELETE("", c.UninstallPluginHandler)
}

func (c *PluginController) StartPluginHandler(ctx *gin.Context) {
	nameStr := ctx.Param("name")
	err := c.service.Start(nameStr)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}
}

func (c *PluginController) StopPluginHandler(ctx *gin.Context) {
	nameStr := ctx.Param("name")
	err := c.service.Stop(nameStr)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}
}

func (c *PluginController) GetPluginHandler(ctx *gin.Context) {
	nameStr := ctx.Param("name")
	ret, err := c.service.GetByName(nameStr)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, ret)
}

func (c *PluginController) GetPluginCommandsHandler(ctx *gin.Context) {
	nameStr := ctx.Param("name")
	result, err := c.service.ListCommand(nameStr)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
	} else {
		ctx.JSON(http.StatusOK, result)
	}
}

func (c *PluginController) ExecPluginCommandHandler(ctx *gin.Context) {
	nameStr := ctx.Param("name")
	commandStr := ctx.Param("command")
	params := ctx.QueryArray("params")

	result, err := c.service.Command(nameStr, commandStr, params...)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
	} else {
		ctx.JSON(http.StatusOK, &shard.PluginCommandExecResult{
			Result: result,
		})
	}
}

func (c *PluginController) ListPluginHandler(ctx *gin.Context) {
	var paging shard.Paging
	if ctx.BindQuery(&paging) != nil {
		return
	}

	total, configurations, err := c.service.ListAll(paging.Page, paging.Size, paging.Sort, paging.Order)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, &shard.PluginsWithCount{
		TotalCount: total,
		Items:      configurations,
	})
}

func (c *PluginController) InstallPluginHandler(ctx *gin.Context) {
	flagStr := ctx.PostForm("flag")
	formFile, err := ctx.FormFile("file")
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, err)
		return
	}
	flag, err := plugin.ParseInstallFlag(flagStr)
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, err)
		return
	}
	file, err := formFile.Open()
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, err)
		return
	}

	result, err := c.service.InstallByReaderAt(file, formFile.Size, flag)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, &result)
}

func (c *PluginController) UninstallPluginHandler(ctx *gin.Context) {
	nameStr := ctx.Query("name")
	flagStr := ctx.Query("flag")
	if nameStr == "" {
		ctx.AbortWithError(http.StatusBadRequest, ErrKeyCannotBeEmpty)
		return
	}
	flag, err := plugin.ParseUninstallFlag(flagStr)
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, err)
		return
	}

	c.service.Uninstall(nameStr, flag)
	ctx.Status(http.StatusOK)
}
