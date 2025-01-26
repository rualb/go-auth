package router

import (
	"fmt"
	"html/template"
	"io"
	"net/http"

	"github.com/labstack/echo-contrib/echoprometheus"
	"github.com/labstack/echo/v4"

	"go-auth/internal/config/consts"
	account "go-auth/internal/controller/account"

	auth "go-auth/internal/controller/auth"
	auth_email "go-auth/internal/controller/auth/email"
	auth_tel "go-auth/internal/controller/auth/tel"

	"go-auth/internal/service"
	xweb "go-auth/internal/web"
	webfs "go-auth/web"

	xlog "go-auth/internal/util/utillog"

	"github.com/labstack/echo/v4/middleware"
)

func Init(e *echo.Echo, appService service.AppService) {

	e.Renderer = mustNewRenderer()

	initCORSConfig(e, appService)

	initAuthController(e, appService)
	initDebugController(e, appService)

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
		xlog.Panic("sys api key is empty")
		return
	}

	if startNewListener {

		e = echo.New() // overwrite override

		e.Use(middleware.Recover())
		// e.Use(middleware.Logger())
	} else {
		xlog.Warn("sys api serve in main listener: %v", listen)
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
			xlog.Info("sys api serve on: %v main: %v", listenSys, listen)

			if err := e.Start(listenSys); err != nil {
				if err != http.ErrServerClosed {
					xlog.Error("%v", err)
				} else {
					xlog.Info("shutting down the server")
				}
			}
		}()

	} else {
		xlog.Info("sys api server serve on main listener: %v", listen)
	}

}

type tmplRenderer struct {
	// viewsMvc  echo.Renderer
	indexHTML *template.Template
}

func (x *tmplRenderer) Render(w io.Writer, name string, data any, c echo.Context) error {

	if name == "index.html" {

		return x.indexHTML.ExecuteTemplate(w, name, data)
	}

	return fmt.Errorf("undef template")
	// return x.viewsMvc.Render(w, name, data, c)

}

func mustNewRenderer() echo.Renderer {

	indexHTML, err := template.New("index.html").Parse(webfs.MustAuthIndexHTML())

	if err != nil {
		panic(err)
	}

	//	err := t.templates.ExecuteTemplate(w, "layout_header", data)

	handler := &tmplRenderer{
		// viewsMvc:  mvc.NewTemplateRenderer(controller.ViewsFs(), "views/auth/*.html"),
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

func initDebugController(e *echo.Echo, _ service.AppService) {

	e.GET(consts.PathAuthPingDebugAPI, func(c echo.Context) error { return c.String(http.StatusOK, "pong") })
	// publicly-available-no-sensitive-data
	e.GET("/health", func(c echo.Context) error { return c.JSON(http.StatusOK, struct{}{}) })

}

func initAuthController(e *echo.Echo, appService service.AppService) {

	appConfig := appService.Config()

	isDebug := appConfig.Debug
	isModeEmail := appConfig.Identity.IsAuthEmail
	isModeTel := appConfig.Identity.IsAuthTel
	IsAuthSignup := appConfig.Identity.IsAuthSignup
	IsAuthForgot := appConfig.Identity.IsAuthForgot

	// type reqHandler func(c echo.Context) error

	// modeAPI := true
	//
	{
		{

			handler := func(c echo.Context) error {
				ctrl := auth.NewIndexController(appService, c)
				return ctrl.Handler()
			}

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

			e.GET(consts.PathAuthAccountSettings, handler,
				xweb.AuthorizeMiddleware(appService, true),
			)

		}

	}

	{

	}

	{
		// Sign up

		if IsAuthSignup {

			if isModeTel {

				handler := func(c echo.Context) error {
					ctrl := auth_tel.NewAccountSignupController(appService, c)
					return ctrl.Handler()
				}

				e.GET(consts.PathAuthSignupTelAPI, func(c echo.Context) error { return handler(c) })
				e.POST(consts.PathAuthSignupTelAPI, func(c echo.Context) error { return handler(c) })

			}

			if isModeEmail {

				handler := func(c echo.Context) error {
					ctrl := auth_email.NewAccountSignupController(appService, c)
					return ctrl.Handler()
				}

				e.GET(consts.PathAuthSignupEmailAPI, func(c echo.Context) error { return handler(c) })
				e.POST(consts.PathAuthSignupEmailAPI, func(c echo.Context) error { return handler(c) })

			}
		}
	}

	{
		// Sign in

		if isModeTel {

			handler := func(c echo.Context) error {
				ctrl := auth_tel.NewAccountSigninController(appService, c)
				return ctrl.Handler()
			}

			e.GET(consts.PathAuthSigninTelAPI, func(c echo.Context) error { return handler(c) })
			e.POST(consts.PathAuthSigninTelAPI, func(c echo.Context) error { return handler(c) })

		}

		if isModeEmail {

			handler := func(c echo.Context) error {
				ctrl := auth_email.NewAccountSigninController(appService, c)
				return ctrl.Handler()
			}

			e.GET(consts.PathAuthSigninEmailAPI, func(c echo.Context) error { return handler(c) })
			e.POST(consts.PathAuthSigninEmailAPI, func(c echo.Context) error { return handler(c) })

		}

	}

	{
		// Forgot password
		if IsAuthForgot {

			if isModeTel {

				handler := func(c echo.Context) error {
					ctrl := auth_tel.NewAccountForgotPasswordController(appService, c)
					return ctrl.Handler()
				}

				e.GET(consts.PathAuthForgotPasswordTelAPI, func(c echo.Context) error { return handler(c) })
				e.POST(consts.PathAuthForgotPasswordTelAPI, func(c echo.Context) error { return handler(c) })

			}

			if isModeEmail {

				handler := func(c echo.Context) error {
					ctrl := auth_email.NewAccountForgotPasswordController(appService, c)
					return ctrl.Handler()
				}

				e.GET(consts.PathAuthForgotPasswordEmailAPI, func(c echo.Context) error { return handler(c) })
				e.POST(consts.PathAuthForgotPasswordEmailAPI, func(c echo.Context) error { return handler(c) })

			}
		}

	}

	{
		// Sign out
		handler := func(c echo.Context) error {
			ctrl := auth.NewAccountSignoutController(appService, c)
			return ctrl.Signout()
		}

		e.GET(consts.PathAuthSignoutAPI, func(c echo.Context) error { return handler(c) })
		e.POST(consts.PathAuthSignoutAPI, func(c echo.Context) error { return handler(c) })

	}
	{
		// Config
		handler := func(c echo.Context) error {
			ctrl := auth.NewConfigAPIController(appService, c)
			return ctrl.Handler()
		}

		e.GET(consts.PathAuthConfigAPI, func(c echo.Context) error { return handler(c) })

	}
	{
		// Status
		handler := func(c echo.Context) error {
			ctrl := auth.NewStatusAPIController(appService, c)
			return ctrl.Handler()
		}

		e.GET(consts.PathAuthStatusAPI, func(c echo.Context) error { return handler(c) },
			xweb.TokenRotateMiddleware(appService), /*rotate auth token*/
		)

	}

	{

		{
			authorize := xweb.AuthorizeMiddleware(appService, false)

			grp := e.Group(consts.PathAuthAccountChangePasswordAPI, authorize)

			handler := func(c echo.Context) error {
				ctrl := account.NewChangePasswordController(appService, c)
				return ctrl.Handler()
			}

			grp.GET("", func(c echo.Context) error { return handler(c) })
			grp.POST("", func(c echo.Context) error { return handler(c) })

		}

		{
			authorize := xweb.AuthorizeMiddleware(appService, false)

			grp := e.Group(consts.PathAuthAccountDeleteDataAPI, authorize)

			handler := func(c echo.Context) error {
				ctrl := account.NewDeleteDataController(appService, c)
				return ctrl.Handler()
			}

			grp.GET("", func(c echo.Context) error { return handler(c) })
			grp.POST("", func(c echo.Context) error { return handler(c) })

		}

	}

}

/////////////////////////////////////////////////////
