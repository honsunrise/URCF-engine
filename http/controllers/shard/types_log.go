package shard

import "github.com/zhsyourai/URCF-engine/models"

type LogWithCount struct {
	TotalCount int64        `json:"total_count"`
	Items      []models.Log `json:"items"`
}
