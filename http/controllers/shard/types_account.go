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

type LoginResponse struct {
	Token string `json:"token"`
}
