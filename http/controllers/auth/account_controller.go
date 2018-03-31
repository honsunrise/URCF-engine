package auth

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/kataras/iris"
	"github.com/zhsyourai/URCF-engine/http/controllers/shard"
	"github.com/zhsyourai/URCF-engine/http/helper"
	"github.com/zhsyourai/URCF-engine/services/account"
)

// AccountController is our /uaa controller.
type AccountController struct {
	Ctx        iris.Context
	Service    account.Service
	JwtHandler *helper.JWT
}

//PutPassword handles PUT:/uaa/password
func (c *AccountController) PutPassword() error {
	changePasswordRequest := &shard.ChangePasswordRequest{}

	if err := c.Ctx.ReadJSON(changePasswordRequest); err != nil {
		return err
	}

	token := c.Ctx.Values().Get("jwt").(*jwt.Token)

	if err := c.Service.ChangePassword(token.Header["id"].(string),
		changePasswordRequest.OldPassword, changePasswordRequest.NewPassword); err != nil {
		return err
	}

	return nil
}
