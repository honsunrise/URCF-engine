package http

import (
	stdContext "context"
	"time"

	"github.com/zhsyourai/URCF-engine/config"
	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/secure"
	"github.com/gin-contrib/cors"
	"net/http"
	"github.com/zhsyourai/URCF-engine/http/gin-jwt"
	"github.com/zhsyourai/URCF-engine/http/controllers"
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

	jwtMiddleware, err := gin_jwt.NewGinJwtMiddleware(gin_jwt.MiddlewareConfig{
		Realm: "urcf",
		SigningAlgorithm: "HS512",
		KeyFunc: func() interface{} {
			return []byte("hahahah")
		},
	})

	if err != nil {
		return err
	}

	jwtGenerator, err := gin_jwt.NewGinJwtGenerator(gin_jwt.GeneratorConfig{
		Issuer: "urcf",
		SigningAlgorithm: "HS512",
		KeyFunc: func() interface{} {
			return []byte("hahahah")
		},
	})

	if err != nil {
		return err
	}

	router.Use(secureConf)
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowCredentials = true
	corsConfig.AllowAllOrigins = true
	router.Use(cors.New(corsConfig))
	// router.Use(jwtHandler.Handler)

	v1 := router.Group("/v1")
	{
		controllers.NewAccountController(jwtMiddleware, jwtGenerator).Handler(v1.Group("/uaa"))
		controllers.NewConfigurationController().Handler(v1.Group("/configuration"))
		controllers.NewLogController().Handler(v1.Group("/log"))
		controllers.NewNetFilterController().Handler(v1.Group("/netfilter"))
		controllers.NewProcessesController().Handler(v1.Group("/process"))
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
