package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/zhsyourai/URCF-engine/services/netfilter"
	"net/http"
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
	root.GET("/list/:table", c.ListChainHandler)
	root.GET("/list/:table/:chain", c.ListHandler)
}

func (c *NetFilterController) ListChainHandler(ctx *gin.Context) {
	tableStr := ctx.Param("table")
	result, err := c.service.ListChains(tableStr)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, &result)
}

func (c *NetFilterController) ListHandler(ctx *gin.Context) {
	tableStr := ctx.Param("table")
	chainStr := ctx.Param("chain")
	result, err := c.service.ListAll(tableStr, chainStr)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, &result)
}
