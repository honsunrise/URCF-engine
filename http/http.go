package http

import (
	stdContext "context"
	"time"

	"github.com/didip/tollbooth"
	corsMiddleware "github.com/iris-contrib/middleware/cors"
	prometheusMiddleware "github.com/iris-contrib/middleware/prometheus"
	"github.com/iris-contrib/middleware/secure"
	"github.com/iris-contrib/middleware/tollboothic"
	"github.com/kataras/iris"
	"github.com/kataras/iris/mvc"
	"github.com/zhsyourai/URCF-engine/http/helper"
	"github.com/zhsyourai/URCF-engine/services/account"
	"github.com/zhsyourai/URCF-engine/http/controllers/anonymous"
	"github.com/zhsyourai/URCF-engine/http/controllers/auth"
)

var (
	app             = iris.New()
	shutdownTimeout = 5 * time.Second
)

const debug = true

func StartHTTPServer() error {
	secureConf := secure.New(secure.Options{
		SSLRedirect:             true,                                            // If SSLRedirect is set to true, then only allow HTTPS requests. Default is false.
		SSLTemporaryRedirect:    false,                                           // If SSLTemporaryRedirect is true, the a 302 will be used while redirecting. Default is false (301).
		SSLProxyHeaders:         map[string]string{"X-Forwarded-Proto": "https"}, // SSLProxyHeaders is set of header keys with associated values that would indicate a valid HTTPS request. Useful when using Nginx: `map[string]string{"X-Forwarded-Proto": "https"}`. Default is blank map.
		STSSeconds:              315360000,                                       // STSSeconds is the max-age of the Strict-Transport-Security header. Default is 0, which would NOT include the header.
		STSIncludeSubdomains:    true,                                            // If STSIncludeSubdomains is set to true, the `includeSubdomains` will be appended to the Strict-Transport-Security header. Default is false.
		STSPreload:              true,                                            // If STSPreload is set to true, the `preload` flag will be appended to the Strict-Transport-Security header. Default is false.
		ForceSTSHeader:          false,                                           // STS header is only included when the connection is HTTPS. If you want to force it to always be added, set to true. `IsDevelopment` still overrides this. Default is false.
		FrameDeny:               true,                                            // If FrameDeny is set to true, adds the X-Frame-Options header with the value of `DENY`. Default is false.
		CustomFrameOptionsValue: "SAMEORIGIN",                                    // CustomFrameOptionsValue allows the X-Frame-Options header value to be set with a custom value. This overrides the FrameDeny option.
		ContentTypeNosniff:      true,                                            // If ContentTypeNosniff is true, adds the X-Content-Type-Options header with the value `nosniff`. Default is false.
		BrowserXSSFilter:        true,                                            // If BrowserXssFilter is true, adds the X-XSS-Protection header with the value `1; mode=block`. Default is false.
		ContentSecurityPolicy:   "default-src 'self'",                            // ContentSecurityPolicy allows the Content-Security-Policy header value to be set with a custom value. Default is "".
		IsDevelopment:           debug,                                           // This will cause the AllowedHosts, SSLRedirect, and STSSeconds/STSIncludeSubdomains options to be ignored during development. When deploying to production, be sure to set this to false.
		IgnorePrivateIPs:        true,
	})

	prometheus := prometheusMiddleware.New("serviceName", 300, 1200, 5000)
	cors := corsMiddleware.New(corsMiddleware.Options{
		AllowedOrigins:   []string{"*"}, // allows everything, use that to change the hosts.
		AllowedMethods:   []string{"*"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
		Debug: debug,
	})

	app.Use(prometheus.ServeHTTP)
	app.Use(secureConf.Serve)
	app.Use(cors)

	err := configureUAA(app)
	if err != nil {
		return err
	}

	return app.Run(iris.Addr(":8080"), iris.WithoutVersionChecker,
		iris.WithoutInterruptHandler, iris.WithConfiguration(iris.YAML("./http/configs/iris.yml")))
}

func configureUAA(app *iris.Application) error {
	jwtHelper, err := helper.NewJWT()
	if err != nil {
		return err
	}
	limiter := tollbooth.NewLimiter(1000, nil)

	mvcApp := mvc.New(app.Party("/uaa", tollboothic.LimitHandler(limiter)))
	mvcApp.Register(
		account.GetInstance(),
		jwtHelper,
	)
	mvcApp.Handle(new(anonymous.AccountController))

	mvcApp = mvc.New(app.Party("/uaa", tollboothic.LimitHandler(limiter), jwtHelper.Serve))
	mvcApp.Register(
		account.GetInstance(),
		jwtHelper,
	)
	mvcApp.Handle(new(auth.AccountController))
	return nil
}

func StopHTTPServer() (err error) {
	ctx, cancel := stdContext.WithTimeout(stdContext.Background(), shutdownTimeout)
	defer cancel()
	// close all hosts
	err = app.Shutdown(ctx)
	return
}
