package shard

import "github.com/zhsyourai/URCF-engine/models"

type PutConfigureRequest struct {
	Key   string      `form:"key" json:"key" binding:"required"`
	Value interface{} `form:"value" json:"value" binding:"required"`
}

type ConfigurationsWithCount struct {
	TotalCount int64           `json:"total_count"`
	Items      []models.Config `json:"items"`
}
