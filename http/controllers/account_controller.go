package controllers

import (
	"github.com/zhsyourai/URCF-engine/services/account"
	"github.com/kataras/iris"
)

type RegisterRequest struct {
}

type RegisterResponse struct {
}

// AccountController is our /uaa controller.
type AccountController struct {
	Ctx iris.Context
	Service account.Service
}

// PostRegister handles POST:/uaa/register.
func (c *AccountController) PostRegister() (RegisterResponse, error) {
	registerRequest := &RegisterRequest{}

	if err := c.Ctx.ReadJSON(registerRequest); err != nil {
		return RegisterResponse{}, err
	}

	//user, err := c.Service.Register()
	//if err != nil {
	//	return RegisterResponse{}, err
	//}

	registerResponse := RegisterResponse{}

	return registerResponse, nil
}

//PutPassword handles PUT:/uaa/password
func (c *AccountController) PutPassword() () {

}

// PostLogin handles POST:/user/login.
func (c *AccountController) PostLogin() {
	var (
		username = c.Ctx.FormValue("username")
		password = c.Ctx.FormValue("password")
	)

	_, err := c.Service.Verify(username, password)
	if err != nil {
		return
	}
}

// AnyLogout handles any method on path /user/logout.
func (c *AccountController) AnyLogout() {
}
