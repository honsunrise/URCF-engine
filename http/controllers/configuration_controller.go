package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/zhsyourai/URCF-engine/services/configuration"
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
}
