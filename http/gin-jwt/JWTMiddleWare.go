package gin_jwt

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
	"github.com/dgrijalva/jwt-go"
	"crypto/rsa"
	"time"
	"errors"
	"crypto/ecdsa"
)

var (
	// ErrMissingRealm indicates Realm name is required
	ErrMissingRealm = errors.New("realm is missing")

	// ErrForbidden when HTTP status 403 is given
	ErrForbidden = errors.New("you don't have permission to access this resource")

	// ErrForbidden when HTTP status 403 is given
	ErrNotSupportSigningAlgorithm = errors.New("this signing algorithm NOT support")

	// ErrExpiredToken indicates JWT token has expired. Can't refresh.
	ErrTokenInvalid = errors.New("token is invalid")

	// ErrEmptyAuthHeader can be thrown if authing with a HTTP header, the Auth header needs to be set
	ErrEmptyAuthHeader = errors.New("auth header is empty")

	// ErrInvalidAuthHeader indicates auth header is invalid, could for example have the wrong Realm name
	ErrInvalidAuthHeader = errors.New("auth header is invalid")

	// ErrEmptyQueryToken can be thrown if authing with URL Query, the query token variable is empty
	ErrEmptyQueryToken = errors.New("query token is empty")

	// ErrEmptyCookieToken can be thrown if authing with a cookie, the token cokie is empty
	ErrEmptyCookieToken = errors.New("cookie token is empty")

	// ErrInvalidSigningAlgorithm indicates signing algorithm is invalid, needs to be HS256, HS384, HS512, RS256, RS384 or RS512
	ErrInvalidSigningAlgorithm = errors.New("invalid signing algorithm")

	// ErrInvalidPrivKey indicates that the given private key is invalid
	ErrInvalidPrivKey = errors.New("private key invalid")

	// ErrInvalidKey indicates the the given public key is invalid
	ErrInvalidKey = errors.New("public key invalid")
)

type Config struct {
	// Realm name to display to the user. Required.
	Realm string

	// signing algorithm - possible values are HS256, HS384, HS512
	// Optional, default is HS256.
	SigningAlgorithm string

	// Secret key used for signing. Required.
	KeyFunc    func() interface{}
	PriKeyFunc func() interface{}

	ErrorHandler func(ctx gin.Context, err error)
	// Duration that a jwt token is valid. Optional, defaults to one hour.
	Timeout time.Duration

	// This field allows clients to refresh their token until MaxRefresh has passed.
	// Note that clients can refresh their token in the last moment of MaxRefresh.
	// This means that the maximum validity timespan for a token is MaxRefresh + Timeout.
	// Optional, defaults to 0 meaning not refreshable.
	MaxRefresh time.Duration
	// TokenLookup is a string in the form of "<source>:<name>" that is used
	// to extract token from the request.
	// Optional. Default value "header:Authorization:Bearer".
	// Possible values:
	// - "header:<name>"
	// - "query:<name>"
	// - "cookie:<name>"
	TokenLookup string

	// NowFunc provides the current time. You can override it to use another time value. This is useful for testing or if your server uses a different time zone than your tokens.
	NowFunc func() time.Time

	ContextKey string
}

type GinJwtHelper struct {
	config Config
	key    interface{}
	priKey interface{}
}

func NewGinJwtHelper(config Config) (*GinJwtHelper, error) {
	if config.Realm == "" {
		return nil, ErrMissingRealm
	}

	helper := &GinJwtHelper{
		config: config,
	}

	if config.TokenLookup == "" {
		config.TokenLookup = "header:Authorization:Bearer"
	}

	if config.SigningAlgorithm == "" {
		config.SigningAlgorithm = "HS256"
	}

	if config.Timeout == 0 {
		config.Timeout = time.Hour
	}

	if config.NowFunc == nil {
		config.NowFunc = time.Now
	}

	switch config.SigningAlgorithm {
	case "RS256", "RS384", "RS512":
		if priKey, ok := config.PriKeyFunc().(*rsa.PrivateKey); ok {
			helper.priKey = priKey
		} else {
			return nil, ErrInvalidPrivKey
		}
		if pubKey, ok := config.KeyFunc().(*rsa.PublicKey); ok {
			helper.key = pubKey
		} else {
			return nil, ErrInvalidKey
		}
	case "EC256", "EC384", "EC512":
		if priKey, ok := config.PriKeyFunc().(*ecdsa.PrivateKey); ok {
			helper.priKey = priKey
		} else {
			return nil, ErrInvalidPrivKey
		}
		if pubKey, ok := config.KeyFunc().(*ecdsa.PublicKey); ok {
			helper.key = pubKey
		} else {
			return nil, ErrInvalidKey
		}
	case "HS256", "HS384", "HS512":
		if key, ok := config.KeyFunc().([]byte); ok {
			helper.key = key
		} else {
			return nil, ErrInvalidKey
		}
	default:
		return nil, ErrNotSupportSigningAlgorithm
	}

	return helper, nil
}

func (h *GinJwtHelper) Middleware(ctx *gin.Context) {
	token, err := h.parseToken(ctx)

	if err != nil {
		ctx.AbortWithError(http.StatusUnauthorized, err)
		return
	}

	err = h.checkJWT(token)

	if err != nil {
		ctx.AbortWithError(http.StatusUnauthorized, err)
		return
	}

	ctx.Set(h.config.ContextKey, token)

	ctx.Next()
}

func (h *GinJwtHelper) parseToken(ctx *gin.Context) (*jwt.Token, error) {
	var token, originToken string
	var err error

	parts := strings.SplitN(h.config.TokenLookup, ":", 3)
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
	if !(len(tmpParts) == 2 && tmpParts[0] == parts [2]) {
		return nil, ErrInvalidAuthHeader
	}

	token = tmpParts[1]

	if err != nil {
		return nil, err
	}

	return jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if jwt.GetSigningMethod(h.config.SigningAlgorithm) != token.Method {
			return nil, ErrInvalidSigningAlgorithm
		}
		return h.key, nil
	})
}

func (h *GinJwtHelper) checkJWT(token *jwt.Token) error {
	if !token.Valid {
		return ErrTokenInvalid
	}
	return nil
}
