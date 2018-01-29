package models

import "time"

type OAuthType int

const (
	GOOGLE   OAuthType = iota
	WECHAT
	QQ
	SINA
	FACEBOOK
	TWITTER
)

type Account struct {
	UID                   string
	Email                 string
	CreateDate            time.Time
	UnionIds              map[OAuthType]string
	Password              string
	AccountNonExpired     bool
	AccountNonLocked      bool
	CredentialsNonExpired bool
	Role                  []string
	Enabled               bool
	UpdateDate            time.Time
}
