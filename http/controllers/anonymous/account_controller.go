package anonymous

import (
	"github.com/zhsyourai/URCF-engine/http/controllers/shard"
	"github.com/zhsyourai/URCF-engine/services/account"
	"net/http"
	"github.com/gin-gonic/gin"
	"time"
	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"crypto/rsa"
	"crypto/rand"
)

func NewAccountController() *AccountController {
	return &AccountController{
		service: account.GetInstance(),
	}
}

// AccountController is our /uaa controller.
type AccountController struct {
	service account.Service
	key     *rsa.PrivateKey
}

func (c *AccountController) Handler(root *gin.RouterGroup) {
	var err error
	c.key, err = rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		panic(err)
	}
	root.GET("/register", c.RegisterHandler)
	root.POST("/login", c.LoginHandler)
	root.POST("/refresh_token", c.RefreshTokenHandler)
}

func (c *AccountController) RegisterHandler(ctx *gin.Context) {

	registerRequest := &shard.RegisterRequest{}
	ctx.Bind(registerRequest)

	user, err := c.service.Register(registerRequest.Username, registerRequest.Password, registerRequest.Roles)
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, err)
	}

	registerResponse := shard.RegisterResponse{
		Username:   user.ID,
		Role:       user.Role,
		CreateDate: user.CreateDate,
	}

	ctx.JSON(http.StatusOK, registerResponse)
}

func (c *AccountController) LoginHandler(ctx *gin.Context) {
	loginRequest := &shard.LoginRequest{}
	ctx.Bind(loginRequest)

	_, err := c.service.Verify(loginRequest.Username, loginRequest.Password)
	if err != nil {
		ctx.AbortWithError(http.StatusUnauthorized, err)
	}

	now := time.Now()
	expire := now.Add(time.Hour * 24)

	// Create the token
	token := jwt.NewWithClaims(jwt.GetSigningMethod("RS512"), jwt.StandardClaims{
		ExpiresAt: expire.Unix(),
		Id:        uuid.Must(uuid.NewRandom()).String(),
		IssuedAt:  now.Unix(),
		Issuer:    "urcf-engine",
		NotBefore: now.Unix(),
	})

	tokenString, err := token.SignedString(c.key)

	ctx.JSON(http.StatusOK, gin.H{
		"access_token":  tokenString,
		"refresh_token": tokenString,
		"expire":        expire.Format(time.RFC3339),
	})
}

func (c *AccountController) RefreshTokenHandler(ctx *gin.Context) {

}
