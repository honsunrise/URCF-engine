package http

import (
	stdContext "context"
	"time"

	"crypto/rand"
	"crypto/rsa"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/secure"
	"github.com/gin-gonic/gin"
	"github.com/zhsyourai/URCF-engine/config"
	"github.com/zhsyourai/URCF-engine/http/controllers"
	"github.com/zhsyourai/URCF-engine/http/gin-jwt"
	"net/http"
)

var (
	s               *http.Server
	shutdownTimeout = 5 * time.Second
)

func StartHTTPServer() error {
	router := gin.Default()
	secureConf := secure.New(secure.Config{
		AllowedHosts:          []string{"example.com", "ssl.example.com"},
		SSLRedirect:           true,
		SSLHost:               "ssl.example.com",
		STSSeconds:            315360000,
		STSIncludeSubdomains:  true,
		FrameDeny:             true,
		ContentTypeNosniff:    true,
		BrowserXssFilter:      true,
		ContentSecurityPolicy: "default-src 'self'",
		IENoOpen:              true,
		ReferrerPolicy:        "strict-origin-when-cross-origin",
		SSLProxyHeaders:       map[string]string{"X-Forwarded-Proto": "https"},
		IsDevelopment:         config.PROD,
	})

	const SigningAlgorithm = "RS512"
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		panic(err)
	}

	jwtMiddleware, err := gin_jwt.NewGinJwtMiddleware(gin_jwt.MiddlewareConfig{
		Realm:            "urcf",
		SigningAlgorithm: SigningAlgorithm,
		KeyFunc: func() interface{} {
			return &key.PublicKey
		},
	})

	if err != nil {
		return err
	}

	jwtGenerator, err := gin_jwt.NewGinJwtGenerator(gin_jwt.GeneratorConfig{
		Issuer:           "urcf",
		SigningAlgorithm: SigningAlgorithm,
		KeyFunc: func() interface{} {
			return key
		},
	})

	if err != nil {
		return err
	}

	router.Use(secureConf)
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowCredentials = true
	corsConfig.AllowAllOrigins = true
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "HEAD", "DELETE"}
	corsConfig.AllowHeaders = []string{"Authorization", "Origin", "Content-Length", "Content-Type"}
	router.Use(cors.New(corsConfig))
	// router.Use(jwtHandler.Handler)

	v1 := router.Group("/v1")
	{
		controllers.NewAccountController(jwtMiddleware, jwtGenerator).Handler(v1.Group("/uaa"))
		controllers.NewConfigurationController().Handler(v1.Group("/configuration"))
		controllers.NewLogController(jwtMiddleware).Handler(v1.Group("/log"))
		controllers.NewNetFilterController().Handler(v1.Group("/netfilter"))
		controllers.NewProcessesController().Handler(v1.Group("/process"))
		controllers.NewPluginController(jwtMiddleware).Handler(v1.Group("/plugin"))
	}

	s = &http.Server{
		Addr:           ":8080",
		Handler:        router,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	return s.ListenAndServe()
}

func StopHTTPServer() (err error) {
	ctx, cancel := stdContext.WithTimeout(stdContext.Background(), shutdownTimeout)
	defer cancel()
	// close all hosts
	err = s.Shutdown(ctx)
	return
}
