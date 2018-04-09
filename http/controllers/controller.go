package controllers

import "github.com/gin-gonic/gin"

type Controller interface {
	Handler(root *gin.RouterGroup)
}
