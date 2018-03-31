package auth

import (
	"github.com/kataras/iris"
	"github.com/zhsyourai/URCF-engine/http/helper"
	"github.com/zhsyourai/URCF-engine/services/account"
)

// ConfigurationController is our /configuration controller.
type ConfigurationController struct {
	Ctx        iris.Context
	Service    account.Service
	JwtHandler *helper.JWT
}
