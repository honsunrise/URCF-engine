package models

import "time"

type GenderType int

const (
	MAN     GenderType = iota
	WOMAN
	UNKNOWN
)

type AccountProfile struct {
	UID string `json:"uid"`
	/**
	 * Account profile
	 */
	Firstname  string     `json:"firstname"`
	Lastname   string     `json:"lastname"`
	AvatarUrl  string     `json:"avatarUrl"`
	Bio        string     `json:"bio"`
	Birthday   time.Time  `json:"birthday"`
	Gender     GenderType `json:"gender"`
	UpdateDate time.Time  `json:"updateDate"`
	Locale     string     `json:"locale"`
}
