package gin_jwt

import "errors"

var (
	// ErrMissingRealm indicates Realm name is required
	ErrMissingRealm = errors.New("realm is missing")

	ErrMissingSigningAlgorithm = errors.New("signing algorithm is missing")

	ErrMissingKeyFunction = errors.New("key function is missing")

	ErrContextNotHaveToken = errors.New("context not have token")

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

	// ErrInvalidKey indicates the the given public key is invalid
	ErrInvalidKey = errors.New("key invalid")
)