package shard

type Paging struct {
	Page  uint64 `json:"page"`
	Size  uint64 `json:"size"`
	Sort  string `json:"sort"`
	Order string `json:"order"`
}
