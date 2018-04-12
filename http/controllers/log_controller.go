package controllers

import (
	"github.com/zhsyourai/URCF-engine/services/log"
	"github.com/gin-gonic/gin"
	"github.com/zhsyourai/URCF-engine/http/gin-jwt"
	"net/http"
	"strconv"
)

func NewLogController(middleware *gin_jwt.JwtMiddleware) *LogController {
	return &LogController{
		service:    log.GetInstance(),
		middleware: middleware,
	}
}

// LogController is our /log controller.
type LogController struct {
	service    log.Service
	middleware *gin_jwt.JwtMiddleware
}

func (c *LogController) Handler(root *gin.RouterGroup) {
	root.GET("/list", c.ListLogHandler)
	root.DELETE("/clean/*id", c.CleanLogHandler)
}

func (c *LogController) ListLogHandler(ctx *gin.Context) {
	//token, err := c.middleware.ExtractToken(ctx)
	//if err != nil {
	//	ctx.AbortWithError(http.StatusUnauthorized, err)
	//	return
	//}
	//claims := token.Claims.(jwt.MapClaims)
	//roles := claims["roles"].([]string)
	//for _, e := range roles {
	//	if e == "admin" {
	//
	//	}
	//}

	logs, err := c.service.ListAll()
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, logs)
}


func (c *LogController) CleanLogHandler(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 0)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.service.Clean(id)
	ctx.Status(http.StatusOK)
}