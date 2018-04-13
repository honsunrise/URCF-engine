package shard

type Paging struct {
	Page  uint32 `json:"page" form:"page" query:"page"`
	Size  uint32 `json:"size" form:"size" query:"size"`
	Sort  string `json:"sort" form:"sort" query:"sort"`
	Order string `json:"order" form:"order" query:"order"`
}
