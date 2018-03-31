package auth

import (
	"github.com/kataras/iris"
	"github.com/zhsyourai/URCF-engine/http/helper"
	"github.com/zhsyourai/URCF-engine/services/account"
)

// LogController is our /log controller.
type LogController struct {
	Ctx        iris.Context
	Service    account.Service
	JwtHandler *helper.JWT
}
