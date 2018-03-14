package auth

import (
	"github.com/kataras/iris"
	"github.com/zhsyourai/URCF-engine/services/account"
	"github.com/zhsyourai/URCF-engine/http/helper"
)

// ProcessesController is our /processes controller.
type ProcessesController struct {
	Ctx     iris.Context
	Service account.Service
	JwtHandler *helper.JWT
}
