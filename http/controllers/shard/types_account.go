package shard

import "time"

type RegisterRequest struct {
	Id       string
	Password string
	Role     []string
}

type RegisterResponse struct {
	Id         string
	Role       []string
	CreateDate time.Time
}

type ChangePasswordRequest struct {
	Id          string
	OldPassword string
	NewPassword string
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	AccessToken string `json:"access_token"`
}
