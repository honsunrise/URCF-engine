package helper

import (
	jwtMiddleware "github.com/iris-contrib/middleware/jwt"
	"github.com/dgrijalva/jwt-go"
	"crypto/rsa"
	"crypto/rand"
	"time"
	"github.com/google/uuid"
)

type JWT struct {
	*jwtMiddleware.Middleware
	key *rsa.PrivateKey
}

func NewJWT() (*JWT, error) {
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		return nil, err
	}
	return &JWT{
		Middleware: jwtMiddleware.New(jwtMiddleware.Config{
			ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
				return &key.PublicKey, nil
			},
			SigningMethod: jwt.SigningMethodRS256,
		}),
		key: key,
	}, nil
}

func (c *JWT) New(id string) (string, error) {
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.StandardClaims{
		Id:        uuid.Must(uuid.NewRandom()).String(),
		ExpiresAt: time.Now().Add(time.Minute * 30).Unix(),
	})
	jwtToken.Header["id"] = id
	return jwtToken.SignedString(c.key)
}
