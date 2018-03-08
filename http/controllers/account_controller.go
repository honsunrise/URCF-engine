package controllers

import (
	"github.com/kataras/iris"
	"github.com/zhsyourai/URCF-engine/services/account"
	"time"
	"github.com/zhsyourai/URCF-engine/http/helper"
)

type RegisterRequest struct {
	Id string
	Password string
	Role []string
}

type RegisterResponse struct {
	Id string
	Role []string
	CreateDate time.Time
}

type ChangePasswordRequest struct {
	Id string
	OldPassword string
	NewPassword string
}

type LoginResponse struct {
	Jwt string
}

// AccountController is our /uaa controller.
type AccountController struct {
	Ctx     iris.Context
	Service account.Service
	JwtHandler *helper.JWT
}

// PostRegister handles POST:/uaa/register.
func (c *AccountController) PostRegister() (RegisterResponse, error) {
	registerRequest := &RegisterRequest{}

	if err := c.Ctx.ReadJSON(registerRequest); err != nil {
		return RegisterResponse{}, err
	}

	user, err := c.Service.Register(registerRequest.Id, registerRequest.Password, registerRequest.Role)
	if err != nil {
		return RegisterResponse{}, err
	}

	registerResponse := RegisterResponse{
		Id: user.ID,
		Role: user.Role,
		CreateDate: user.CreateDate,
	}

	return registerResponse, nil
}

//PutPassword handles PUT:/uaa/password
func (c *AccountController) PutPassword() error {
	changePasswordRequest := &ChangePasswordRequest{}

	if err := c.Ctx.ReadJSON(changePasswordRequest); err != nil {
		return err
	}

	// TODO: add server method

	return nil
}

// PostLogin handles POST:/user/login.
func (c *AccountController) PostLogin() (LoginResponse, error) {
	var (
		username = c.Ctx.FormValue("username")
		password = c.Ctx.FormValue("password")
	)

	_, err := c.Service.Verify(username, password)
	if err != nil {
		return LoginResponse{}, err
	}

	token, err := c.JwtHandler.New()
	if err != nil {
		return LoginResponse{}, err
	}
	return LoginResponse{
		Jwt: token,
	}, nil
}

// AnyLogout handles any method on path /user/logout.
func (c *AccountController) AnyLogout() {
}
