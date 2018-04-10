package controllers

import (
	"github.com/zhsyourai/URCF-engine/services/configuration"
	"github.com/gin-gonic/gin"
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
