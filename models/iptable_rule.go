package models

type IptableRule struct {
	Table string `json:"table"`
	Chain string `json:"chain"`
}
