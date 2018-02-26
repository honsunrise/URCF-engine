package shared

type RegisterParam struct {
	Id string
	Password string
	Role []string
}

type VerifyParam struct {
	Id string
	Password string
}