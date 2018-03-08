package http

import (
	stdContext "context"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/didip/tollbooth"
	"github.com/iris-contrib/middleware/cors"
	prometheusMiddleware "github.com/iris-contrib/middleware/prometheus"
	"github.com/iris-contrib/middleware/secure"
	"github.com/iris-contrib/middleware/tollboothic"
	"github.com/kataras/iris"
	"github.com/kataras/iris/mvc"
	"github.com/zhsyourai/URCF-engine/http/helper"
	"github.com/zhsyourai/URCF-engine/services/account"
	"github.com/zhsyourai/URCF-engine/http/controllers"
)

func myHandler(ctx iris.Context) {
	user := ctx.Values().Get("jwt").(*jwt.Token)

	ctx.Writef("This is an authenticated request\n")
	ctx.Writef("Claim content:\n")

	ctx.Writef("%s", user.Signature)
}

var (
	app             = iris.New()
	shutdownTimeout = 5 * time.Second
)

const debug = false

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
	crs := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"}, // allows everything, use that to change the hosts.
		AllowCredentials: true,
	})

	limiter := tollbooth.NewLimiter(1000, nil)

	app.Use(prometheus.ServeHTTP)
	app.Use(secureConf.Serve)
	app.Use(crs)

	err := configureUAA(mvc.New(app.Party("/uaa", tollboothic.LimitHandler(limiter), myHandler)))
	if err != nil {
		return err
	}

	iris.RegisterOnInterrupt(func() {
		ctx, cancel := stdContext.WithTimeout(stdContext.Background(), shutdownTimeout)
		defer cancel()
		// close all hosts
		app.Shutdown(ctx)
	})

	return app.Run(iris.Addr(":8080"), iris.WithoutVersionChecker,
		iris.WithoutInterruptHandler, iris.WithConfiguration(iris.YAML("./http/configs/iris.yml")))
}

func configureUAA(app *mvc.Application) error {
	jwtHandler, err := helper.NewJWT()
	if err != nil {
		return err
	}
	// any dependencies bindings here...
	app.Register(
		account.GetInstance(),
		jwtHandler,
	)

	// controllers registration here...
	app.Handle(new(controllers.AccountController))
	return nil
}

func StopHTTPServer() (err error) {
	ctx, cancel := stdContext.WithTimeout(stdContext.Background(), shutdownTimeout)
	defer cancel()
	// close all hosts
	err = app.Shutdown(ctx)
	return
}
