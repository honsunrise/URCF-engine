package shard

import "time"

type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type RegisterRequest struct {
	Username string   `form:"username" json:"username" binding:"required"`
	Password string   `form:"password" json:"password" binding:"required"`
	Roles    []string `form:"roles" json:"roles" binding:"required"`
}

type RegisterResponse struct {
	Username   string    `json:"username"`
	Role       []string  `json:"role"`
	CreateDate time.Time `json:"create_date"`
}

type ChangePasswordRequest struct {
	Id          string
	OldPassword string
	NewPassword string
}

type LoginRequest struct {
	Username string `form:"username" json:"username" binding:"required"`
	Password string `form:"password" json:"password" binding:"required"`
}

type LoginResponse struct {
	AccessToken string `json:"access_token"`
}
