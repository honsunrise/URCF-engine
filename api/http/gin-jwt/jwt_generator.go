package gin_jwt

import (
	"crypto/ecdsa"
	"crypto/rsa"
	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"time"
)

type GeneratorConfig struct {
	Issuer string

	SigningAlgorithm string

	KeyFunc func() interface{}

	NowFunc func() time.Time
}

type JwtGenerator struct {
	config GeneratorConfig
	key    interface{}
}

func NewGinJwtGenerator(config GeneratorConfig) (*JwtGenerator, error) {
	if config.SigningAlgorithm == "" {
		return nil, ErrMissingSigningAlgorithm
	}

	if config.KeyFunc == nil {
		return nil, ErrMissingKeyFunction
	}

	if config.NowFunc == nil {
		config.NowFunc = time.Now
	}

	var realKey interface{}
	switch config.SigningAlgorithm {
	case "RS256", "RS384", "RS512":
		if pubKey, ok := config.KeyFunc().(*rsa.PrivateKey); ok {
			realKey = pubKey
		} else {
			return nil, ErrInvalidKey
		}
	case "EC256", "EC384", "EC512":
		if pubKey, ok := config.KeyFunc().(*ecdsa.PrivateKey); ok {
			realKey = pubKey
		} else {
			return nil, ErrInvalidKey
		}
	case "HS256", "HS384", "HS512":
		if key, ok := config.KeyFunc().([]byte); ok {
			realKey = key
		} else {
			return nil, ErrInvalidKey
		}
	default:
		return nil, ErrNotSupportSigningAlgorithm
	}

	return &JwtGenerator{
		config: config,
		key:    realKey,
	}, nil
}

func (g *JwtGenerator) GenerateJwt(timeout time.Duration,
	refresh time.Duration, claims map[string]interface{}) (string, error) {
	now := g.config.NowFunc()
	expire := now.Add(timeout)
	finalClaims := jwt.MapClaims{
		"exp": expire.Unix(),
		"jti": uuid.Must(uuid.NewRandom()).String(),
		"iat": now.Unix(),
		"iss": g.config.Issuer,
		"nbf": now.Unix(),
	}

	for key := range claims {
		finalClaims[key] = claims[key]
	}

	// Create the token
	token := jwt.NewWithClaims(jwt.GetSigningMethod(g.config.SigningAlgorithm), finalClaims)

	tokenString, err := token.SignedString(g.key)
	if err != nil {
		return "", err
	} else {
		return tokenString, nil
	}
}
