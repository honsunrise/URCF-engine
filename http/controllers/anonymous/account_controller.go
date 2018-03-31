package anonymous

import (
	"github.com/kataras/iris"
	log "github.com/sirupsen/logrus"
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

// PostRegister handles POST:/uaa/register.
func (c *AccountController) PostRegister() (shard.RegisterResponse, error) {
	registerRequest := &shard.RegisterRequest{}

	if err := c.Ctx.ReadJSON(registerRequest); err != nil {
		return shard.RegisterResponse{}, err
	}

	user, err := c.Service.Register(registerRequest.Id, registerRequest.Password, registerRequest.Role)
	if err != nil {
		return shard.RegisterResponse{}, err
	}

	registerResponse := shard.RegisterResponse{
		Id:         user.ID,
		Role:       user.Role,
		CreateDate: user.CreateDate,
	}

	return registerResponse, nil
}

// PostLogin handles POST:/uaa/login.
func (c *AccountController) PostLogin() (shard.LoginResponse, error) {
	var (
		username = c.Ctx.FormValue("username")
		password = c.Ctx.FormValue("password")
	)

	acc, err := c.Service.Verify(username, password)
	if err != nil {
		return shard.LoginResponse{}, err
	}

	token, err := c.JwtHandler.New(acc.ID)
	if err != nil {
		return shard.LoginResponse{}, err
	}
	return shard.LoginResponse{
		Token: token,
	}, nil
}

// AnyLogout handles any method on path /uaa/logout.
func (c *AccountController) AnyLogout() {
	log.Info("Logout")
}
