package gin_jwt

import (
	"crypto/ecdsa"
	"crypto/rsa"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

const DEFAULT_CONTEXT_KEY = "_JWT_TOKEN_"

type MiddlewareConfig struct {
	Realm string

	SigningAlgorithm string

	KeyFunc func() interface{}

	ErrorHandler func(ctx *gin.Context, err error)

	TokenLookup string

	ContextKey string
}

type JwtMiddleware struct {
	config MiddlewareConfig
	key    interface{}
	priKey interface{}
}

func NewGinJwtMiddleware(config MiddlewareConfig) (*JwtMiddleware, error) {
	if config.Realm == "" {
		return nil, ErrMissingRealm
	}

	if config.SigningAlgorithm == "" {
		return nil, ErrMissingSigningAlgorithm
	}

	if config.KeyFunc == nil {
		return nil, ErrMissingKeyFunction
	}

	if config.TokenLookup == "" {
		config.TokenLookup = "header:Authorization:Bearer"
	}

	if config.ErrorHandler == nil {
		config.ErrorHandler = func(ctx *gin.Context, err error) {
			ctx.Header("WWW-Authenticate", "JWT realm="+config.Realm)
			ctx.AbortWithError(http.StatusUnauthorized, err)
		}
	}

	if config.ContextKey == "" {
		config.ContextKey = DEFAULT_CONTEXT_KEY
	}

	var realKey interface{}
	switch config.SigningAlgorithm {
	case "RS256", "RS384", "RS512":
		if pubKey, ok := config.KeyFunc().(*rsa.PublicKey); ok {
			realKey = pubKey
		} else {
			return nil, ErrInvalidKey
		}
	case "EC256", "EC384", "EC512":
		if pubKey, ok := config.KeyFunc().(*ecdsa.PublicKey); ok {
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

	return &JwtMiddleware{
		config: config,
		key:    realKey,
	}, nil
}

func (m *JwtMiddleware) Handler(ctx *gin.Context) {
	token, err := m.extractToken(ctx)

	if err != nil {
		m.config.ErrorHandler(ctx, err)
		return
	}

	err = m.checkToken(token)

	if err != nil {
		m.config.ErrorHandler(ctx, err)
		return
	}

	ctx.Set(m.config.ContextKey, token)

	ctx.Next()
}

func (m *JwtMiddleware) ExtractToken(ctx *gin.Context) (*jwt.Token, error) {
	if token, ok := ctx.Get(m.config.ContextKey); ok {
		return token.(*jwt.Token), nil
	} else {
		return nil, ErrContextNotHaveToken
	}
}

func (m *JwtMiddleware) extractToken(ctx *gin.Context) (*jwt.Token, error) {
	var token, originToken string
	var err error

	parts := strings.SplitN(m.config.TokenLookup, ":", 3)
	switch parts[0] {
	case "header":
		originToken = ctx.Request.Header.Get(parts[1])
		if originToken == "" {
			return nil, ErrEmptyAuthHeader
		}
	case "query":
		originToken = ctx.Query(parts[1])
		if originToken == "" {
			return nil, ErrEmptyQueryToken
		}
	case "cookie":
		originToken, _ = ctx.Cookie(parts[1])
		if originToken == "" {
			return nil, ErrEmptyCookieToken
		}
	}

	tmpParts := strings.SplitN(originToken, " ", 2)
	if !(len(tmpParts) == 2 && tmpParts[0] == parts[2]) {
		return nil, ErrInvalidAuthHeader
	}

	token = tmpParts[1]

	if err != nil {
		return nil, err
	}

	return jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if jwt.GetSigningMethod(m.config.SigningAlgorithm) != token.Method {
			return nil, ErrInvalidSigningAlgorithm
		}
		return m.key, nil
	})
}

func (m *JwtMiddleware) checkToken(token *jwt.Token) error {
	if !token.Valid {
		return ErrTokenInvalid
	}
	return nil
}
