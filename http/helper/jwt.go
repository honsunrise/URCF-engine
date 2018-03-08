package helper

import (
	jwtMiddleware "github.com/iris-contrib/middleware/jwt"
	"github.com/dgrijalva/jwt-go"
	"crypto/rsa"
	"crypto/rand"
)

type JWT struct {
	key *rsa.PrivateKey
}

func NewJWT() (*JWT, error) {
	key, err := rsa.GenerateKey(rand.Reader, 256)
	if err != nil {
		return nil, err
	}
	return &JWT{
		key: key,
	}, nil
}

func (c *JWT) Handler() *jwtMiddleware.Middleware {
	return jwtMiddleware.New(jwtMiddleware.Config{
		ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
			return c.key, nil
		},
		SigningMethod: jwt.SigningMethodRS256,
	})
}

func (c *JWT) New() (string, error) {
	return jwt.New(jwt.SigningMethodRS256).SignedString(c.key)
}