package controllers

import (
	"github.com/kataras/iris/mvc"
	"github.com/zhsyourai/teddy-backend/uaa/services"
	"github.com/kataras/iris"
)

type RegisterRequest struct {
}

type RegisterResponse struct {
}

// AccountController is our /uaa controller.
type AccountController struct {
	mvc.C
	Service services.AccountService
}

// PostRegister handles POST:/uaa/register.
func (c *AccountController) PostRegister() (RegisterResponse, error) {
	registerRequest := &RegisterRequest{}

	if err := c.Ctx.ReadJSON(registerRequest); err != nil {
		return RegisterResponse{}, err
	}

	user, err := c.Service.CreateUser()
	if err != nil {
		return RegisterResponse{}, err
	}

	registerResponse := RegisterResponse{}

	return registerResponse, nil
}

//DeletePassword handles DELETE:/uaa/password
func (c *AccountController) DeletePassword() () {

}

//PutPassword handles PUT:/uaa/password
func (c *AccountController) PutPassword() () {

}

// GetLogin handles GET:/user/login.
func (c *AccountController) GetLogin() () {
	if c.isLoggedIn() {
		c.logout()
		return
	}
	c.Data["Title"] = "User Login"
	c.Tmpl = PathLogin + ".html"
}

// PostLogin handles POST:/user/login.
func (c *AccountController) PostLogin() {
	var (
		username = c.Ctx.FormValue("username")
		password = c.Ctx.FormValue("password")
	)

	user, err := c.verify(username, password)
	if err != nil {
		c.fireError(err)
		return
	}

	c.Session.Set(sessionIDKey, user.ID)
	c.Path = pathMyProfile
}

// AnyLogout handles any method on path /user/logout.
func (c *AccountController) AnyLogout() {
	c.logout()
}

// GetMe handles GET:/user/me.
func (c *AccountController) GetMe() {
	id, err := c.Session.GetInt64(sessionIDKey)
	if err != nil || id <= 0 {
		// when not already logged in.
		c.Path = PathLogin
		return
	}

	u, found := c.Source.GetByID(id)
	if !found {
		// if the  session exists but for some reason the user doesn't exist in the "database"
		// then logout him and redirect to the register page.
		c.logout()
		return
	}

	// set the model and render the view template.
	c.User = u
	c.Data["Title"] = "Profile of " + u.Username
	c.Tmpl = pathMyProfile + ".html"
}


// GetBy handles GET:/user/{id:long},
// i.e http://localhost:8080/user/1
func (c *AccountController) GetBy(userID int64) {
	// we have /user/{id}
	// fetch and render user json.
	if user, found := c.Source.GetByID(userID); !found {
		// not user found with that ID.
		c.renderNotFound(userID)
	} else {
		c.Ctx.JSON(user)
	}
}
