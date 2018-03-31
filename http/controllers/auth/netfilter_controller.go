package auth

import (
	"github.com/kataras/iris"
	"github.com/zhsyourai/URCF-engine/http/helper"
	"github.com/zhsyourai/URCF-engine/services/account"
)

// NetFilterController is our /netfilter controller.
type NetFilterController struct {
	Ctx        iris.Context
	Service    account.Service
	JwtHandler *helper.JWT
}
