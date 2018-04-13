package shared

type RegisterParam struct {
	Username string
	Password string
	Role     []string
}

type VerifyParam struct {
	Username string
	Password string
}
