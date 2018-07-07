package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/kataras/iris/core/errors"
	"github.com/zhsyourai/URCF-engine/api/http/controllers/shard"
	"github.com/zhsyourai/URCF-engine/services/configuration"
	"net/http"
)

var (
	ErrKeyCannotBeEmpty = errors.New("key can't be empty")
)

func NewConfigurationController() *ConfigurationController {
	return &ConfigurationController{
		service: configuration.GetInstance(),
	}
}

// ConfigurationController is our /configuration controller.
type ConfigurationController struct {
	service configuration.Service
}

func (c *ConfigurationController) Handler(root *gin.RouterGroup) {
	root.GET("/list", c.ListConfigurationHandler)
	root.GET("/", c.GetConfigurationHandler)
	root.PUT("/", c.UpdateConfigurationHandler)
	root.DELETE("/", c.DeleteConfigurationHandler)
}

func (c *ConfigurationController) GetConfigurationHandler(ctx *gin.Context) {
	keyStr := ctx.Query("key")
	log, err := c.service.Get(keyStr)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, log)
}

func (c *ConfigurationController) ListConfigurationHandler(ctx *gin.Context) {
	var paging shard.Paging
	if ctx.BindQuery(&paging) != nil {
		return
	}

	total, configurations, err := c.service.ListAll(paging.Page, paging.Size, paging.Sort, paging.Order)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, &shard.ConfigurationsWithCount{
		TotalCount: total,
		Items:      configurations,
	})
}

func (c *ConfigurationController) UpdateConfigurationHandler(ctx *gin.Context) {
	request := &shard.PutConfigureRequest{}
	ctx.Bind(request)
	if request.Key == "" {
		ctx.AbortWithError(http.StatusBadRequest, ErrKeyCannotBeEmpty)
		return
	}
	c.service.Put(request.Key, request.Value)
	ctx.Status(http.StatusOK)
}

func (c *ConfigurationController) DeleteConfigurationHandler(ctx *gin.Context) {
	keyStr := ctx.Query("key")
	if keyStr == "" {
		ctx.AbortWithError(http.StatusBadRequest, ErrKeyCannotBeEmpty)
		return
	}
	c.service.Delete(keyStr)
	ctx.Status(http.StatusOK)
}
