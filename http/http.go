package http

import (
	stdContext "context"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
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
	api             *http.Server
	webs            *http.Server
	w               errgroup.Group
	shutdownTimeout = 10 * time.Second
)

func apiServer() (*http.Server, error) {
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
		return nil, err
	}

	jwtGenerator, err := gin_jwt.NewGinJwtGenerator(gin_jwt.GeneratorConfig{
		Issuer:           "urcf",
		SigningAlgorithm: SigningAlgorithm,
		KeyFunc: func() interface{} {
			return key
		},
	})

	if err != nil {
		return nil, err
	}

	router.Use(secureConf)
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowCredentials = true
	corsConfig.AllowAllOrigins = true
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "HEAD", "DELETE"}
	corsConfig.AllowHeaders = []string{"Authorization", "Origin", "Content-Length", "Content-Type"}
	router.Use(cors.New(corsConfig))

	v1 := router.Group("/v1")
	{
		controllers.NewAccountController(jwtMiddleware, jwtGenerator).Handler(v1.Group("/uaa"))
		controllers.NewConfigurationController().Handler(v1.Group("/configuration"))
		controllers.NewLogController(jwtMiddleware).Handler(v1.Group("/log"))
		controllers.NewNetFilterController().Handler(v1.Group("/netfilter"))
		controllers.NewProcessesController().Handler(v1.Group("/process"))
		controllers.NewPluginController(jwtMiddleware).Handler(v1.Group("/plugins"))
	}

	return &http.Server{
		Addr:           ":8080",
		Handler:        router,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}, nil
}

func websServer() (*http.Server, error) {
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
		return nil, err
	}

	if err != nil {
		return nil, err
	}

	router.Use(secureConf)
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowCredentials = true
	corsConfig.AllowAllOrigins = true
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "HEAD", "DELETE"}
	corsConfig.AllowHeaders = []string{"Authorization", "Origin", "Content-Length", "Content-Type"}
	router.Use(cors.New(corsConfig))

	controllers.NewWebsController(jwtMiddleware).Handler(router.Group("/"))

	return &http.Server{
		Addr:           ":8081",
		Handler:        router,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}, nil
}

func StartHTTPServer() error {
	api, err := apiServer()
	if err != nil {
		log.Fatal(err)
	}

	webs, err := websServer()
	if err != nil {
		log.Fatal(err)
	}

	w.Go(func() error {
		err := api.ListenAndServe()
		if err != nil {
			log.Fatal(err)
		}
		return err
	})

	w.Go(func() error {
		err := webs.ListenAndServe()
		if err != nil {
			log.Fatal(err)
		}
		return err
	})

	return w.Wait()
}

func StopHTTPServer() (err error) {
	ctx, cancel := stdContext.WithTimeout(stdContext.Background(), shutdownTimeout)
	defer cancel()

	err = api.Shutdown(ctx)
	if err != nil {
		log.Fatal(err)
	}
	err = webs.Shutdown(ctx)
	if err != nil {
		log.Fatal(err)
	}
	return
}
