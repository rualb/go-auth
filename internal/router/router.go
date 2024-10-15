package router

import (
	"html/template"
	"io"
	"net/http"

	"github.com/labstack/echo-contrib/echoprometheus"
	"github.com/labstack/echo/v4"

	"go-auth/internal/config/consts"
	controller "go-auth/internal/controller"

	auth "go-auth/internal/controller/auth"
	auth_email "go-auth/internal/controller/auth/email"
	auth_phonenumber "go-auth/internal/controller/auth/phonenumber"

	mvc "go-auth/internal/mvc"
	"go-auth/internal/service"
	xweb "go-auth/internal/web"
	webfs "go-auth/web"

	xlog "go-auth/internal/tool/toollog"

	"github.com/labstack/echo/v4/middleware"
)

func Init(e *echo.Echo, appService service.AppService) {

	e.Renderer = mustNewRenderer()

	initCORSConfig(e, appService)

	initTestController(e, appService)
	initAuthController(e, appService)
	initHealthController(e, appService)

	initSys(e, appService)
}

func initSys(e *echo.Echo, appService service.AppService) {

	// !!! DANGER for private(non-public) services only
	// or use non-public port via echo.New()

	appConfig := appService.Config()

	listen := appConfig.HTTPServer.Listen
	listenSys := appConfig.HTTPServer.ListenSys
	sysMetrics := appConfig.HTTPServer.SysMetrics
	hasAnyService := sysMetrics
	sysAPIKey := appConfig.HTTPServer.SysAPIKey
	hasAPIKey := sysAPIKey != ""
	hasListenSys := listenSys != ""
	startNewListener := listenSys != listen

	if !hasListenSys {
		return
	}

	if !hasAnyService {
		return
	}

	if !hasAPIKey {
		xlog.Panic("Sys api key is empty")
		return
	}

	if startNewListener {

		e = echo.New() // overwrite override

		e.Use(middleware.Recover())
		// e.Use(middleware.Logger())
	} else {
		xlog.Warn("Sys api serve in main listener: %v", listen)
	}

	sysAPIAccessAuthMW := middleware.KeyAuthWithConfig(middleware.KeyAuthConfig{
		KeyLookup: "query:api-key,header:Authorization",
		Validator: func(key string, c echo.Context) (bool, error) {
			return key == sysAPIKey, nil
		},
	})

	if sysMetrics {
		// may be eSys := echo.New() // this Echo will run on separate port
		e.GET(
			consts.PathSysMetricsAPI,
			echoprometheus.NewHandler(),
			sysAPIAccessAuthMW,
		) // adds route to serve gathered metrics

	}

	if startNewListener {

		// start as async task
		go func() {
			xlog.Info("Sys api serve on: %v main: %v", listenSys, listen)

			if err := e.Start(listenSys); err != nil {
				if err != http.ErrServerClosed {
					xlog.Error("%v", err)
				} else {
					xlog.Info("shutting down the server")
				}
			}
		}()

	} else {
		xlog.Info("Sys api server serve on main listener: %v", listen)
	}

}

type tmplRenderer struct {
	viewsMvc  echo.Renderer
	indexHTML *template.Template
}

func (x *tmplRenderer) Render(w io.Writer, name string, data any, c echo.Context) error {

	if name == "auth.html" {

		return x.indexHTML.ExecuteTemplate(w, name, data)
	}

	return x.viewsMvc.Render(w, name, data, c)

}

func mustNewRenderer() echo.Renderer {

	indexHTML, err := template.New("auth.html").Parse(webfs.MustAuthIndexHTML())

	if err != nil {
		panic(err)
	}

	//	err := t.templates.ExecuteTemplate(w, "layout_header", data)

	handler := &tmplRenderer{
		viewsMvc:  mvc.NewTemplateRenderer(controller.ViewsFs(), "views/auth/*.html"),
		indexHTML: indexHTML,
	}

	return handler

}

func initCORSConfig(e *echo.Echo, _ service.AppService) {

	// CorsEnabled := true
	// if CorsEnabled {
	// 	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
	// 		AllowCredentials:                         true,
	// 		UnsafeWildcardOriginWithAllowCredentials: true,
	// 		AllowOrigins:                             []string{"*"},
	// 		MaxAge:                                   86400,
	// 	}))
	// }
}

func initHealthController(e *echo.Echo, appService service.AppService) {
	ping := controller.NewPingController(appService)
	e.GET(consts.PathAuthTestPingAPI, func(c echo.Context) error { return ping.Ping(c) })
}
func initTestController(e *echo.Echo, appService service.AppService) {

	//

	{

		handler := func(c echo.Context) error {
			ctrl := auth.NewHelloWorldController(appService, c)
			return ctrl.Handler()
		}

		e.GET(consts.PathAuthHelloWorld, handler)

	}

}

func initAuthController(e *echo.Echo, appService service.AppService) {

	appConfig := appService.Config()

	isDebug := appConfig.Debug
	isModeEmail := appConfig.Identity.IsAuthEmail
	isModePhone := appConfig.Identity.IsAuthPhone
	IsAuthSignup := appConfig.Identity.IsAuthSignup
	IsAuthForgot := appConfig.Identity.IsAuthForgot

	// type reqHandler func(c echo.Context, isAPIMode bool) error

	// modeAPI := true
	//
	{
		{

			handler := func(c echo.Context) error {
				ctrl := auth.NewIndexController(appService, c)
				return ctrl.Handler()
			}

			// e.GET(consts.PathAuthLockout, handler)
			// e.GET(consts.PathAuthAccessDenied, handler)

			if isDebug {
				e.GET(consts.PathAuthHelloWorld, handler)
			}

			if IsAuthSignup {
				e.GET(consts.PathAuthSignup, handler)
			}

			e.GET(consts.PathAuthSignin, handler)

			if IsAuthForgot {
				e.GET(consts.PathAuthForgotPassword, handler)
			}

			e.GET(consts.PathAuthSignout, handler)
			e.GET(consts.PathAuthLockout, handler)
			e.GET(consts.PathAuthAccessDenied, handler)
		}

	}
	// {
	// 	handler := func(c echo.Context) error {
	// 		ctrl := account.NewLockoutController(appService)
	// 		return ctrl.Handler(c)
	// 	}

	// 	e.GET(consts.PathAuthLockout, handler)

	// }

	// {

	// 	handler := func(c echo.Context) error {
	// 		ctrl := account.NewAccessDeniedController(appService)
	// 		return ctrl.Handler(c)
	// 	}

	// 	e.GET(consts.PathAuthAccessDenied, handler)

	// }

	{

	}

	{
		// Sign up

		if IsAuthSignup {

			if isModePhone {

				handler := func(c echo.Context, isAPIMode bool) error {
					ctrl := auth_phonenumber.NewAccountSignupController(appService, c, isAPIMode)
					return ctrl.Handler()
				}

				e.GET(consts.PathAuthSignupPhoneNumberAPI, func(c echo.Context) error { return handler(c, true) })
				e.POST(consts.PathAuthSignupPhoneNumberAPI, func(c echo.Context) error { return handler(c, true) })

			}

			if isModeEmail {

				handler := func(c echo.Context, isAPIMode bool) error {
					ctrl := auth_email.NewAccountSignupController(appService, c, isAPIMode)
					return ctrl.Handler()
				}

				e.GET(consts.PathAuthSignupEmailAPI, func(c echo.Context) error { return handler(c, true) })
				e.POST(consts.PathAuthSignupEmailAPI, func(c echo.Context) error { return handler(c, true) })

			}
		}
	}

	{
		// Sign in

		if isModePhone {

			handler := func(c echo.Context, isAPIMode bool) error {
				ctrl := auth_phonenumber.NewAccountSigninController(appService, c, isAPIMode)
				return ctrl.Handler()
			}

			e.GET(consts.PathAuthSigninPhoneNumberAPI, func(c echo.Context) error { return handler(c, true) })
			e.POST(consts.PathAuthSigninPhoneNumberAPI, func(c echo.Context) error { return handler(c, true) })

		}

		if isModeEmail {

			handler := func(c echo.Context, isAPIMode bool) error {
				ctrl := auth_email.NewAccountSigninController(appService, c, isAPIMode)
				return ctrl.Handler()
			}

			e.GET(consts.PathAuthSigninEmailAPI, func(c echo.Context) error { return handler(c, true) })
			e.POST(consts.PathAuthSigninEmailAPI, func(c echo.Context) error { return handler(c, true) })

		}

		// e.GET(prefix, func(c echo.Context) error { return handler(c, false) })
		// e.POST(prefix, func(c echo.Context) error { return handler(c, false) })

	}

	{
		// Forgot password
		if IsAuthForgot {

			if isModePhone {

				handler := func(c echo.Context, isAPIMode bool) error {
					ctrl := auth_phonenumber.NewAccountForgotPasswordController(appService, c, isAPIMode)
					return ctrl.Handler()
				}

				e.GET(consts.PathAuthForgotPasswordPhoneNumberAPI, func(c echo.Context) error { return handler(c, true) })
				e.POST(consts.PathAuthForgotPasswordPhoneNumberAPI, func(c echo.Context) error { return handler(c, true) })

			}

			if isModeEmail {

				handler := func(c echo.Context, isAPIMode bool) error {
					ctrl := auth_email.NewAccountForgotPasswordController(appService, c, isAPIMode)
					return ctrl.Handler()
				}

				e.GET(consts.PathAuthForgotPasswordEmailAPI, func(c echo.Context) error { return handler(c, true) })
				e.POST(consts.PathAuthForgotPasswordEmailAPI, func(c echo.Context) error { return handler(c, true) })

			}
		}
		// e.GET(prefix, func(c echo.Context) error { return handler(c, false) })
		// e.POST(prefix, func(c echo.Context) error { return handler(c, false) })

	}

	{
		// Sign out
		handler := func(c echo.Context, isAPIMode bool) error {
			ctrl := auth.NewAccountSignoutController(appService, c, isAPIMode)
			return ctrl.Signout()
		}

		e.GET(consts.PathAuthSignoutAPI, func(c echo.Context) error { return handler(c, true) })
		e.POST(consts.PathAuthSignoutAPI, func(c echo.Context) error { return handler(c, true) })

	}
	{
		// Info
		handler := func(c echo.Context, isAPIMode bool) error {
			ctrl := auth.NewAccountInfoController(appService, c, isAPIMode)
			return ctrl.Info()
		}

		e.GET(consts.PathAuthStatusAPI, func(c echo.Context) error { return handler(c, true) },
			xweb.TokenRotateMiddleware(appService), /*rotate auth token*/
		)

	}
	{

		e.GET(consts.PathAuthManager, func(c echo.Context) error {

			return c.String(http.StatusOK, "manager")

		}, xweb.AuthorizeMiddleware(appService, true))

	}
}

/////////////////////////////////////////////////////
