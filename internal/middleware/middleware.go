package middleware

import (
	"go-auth/internal/config/consts"
	"go-auth/internal/service"
	xweb "go-auth/internal/web"
	"io/fs"

	xlog "go-auth/internal/tool/toollog"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/labstack/echo-contrib/echoprometheus"
)

func Init(e *echo.Echo, appService service.AppService) {

	appConfig := appService.Config()

	e.HTTPErrorHandler = newHTTPErrorHandler(appService)

	e.Use(middleware.Recover()) //!!!

	if appConfig.HTTPServer.AccessLog {
		e.Use(middleware.Logger())
	}

	e.Use(middleware.Gzip())
	//
	e.Use(xweb.UserLangMiddleware(appService))
	e.Use(xweb.TokenParserMiddleware(appService))

	//
	e.Use(xweb.CsrfMiddleware(appService))

	initSys(e, appService)
}

func initSys(e *echo.Echo, appService service.AppService) {

	appConfig := appService.Config()

	// name := "" // appConfig.Name // name as var

	if appConfig.HTTPServer.Metrics {
		e.Use(echoprometheus.NewMiddlewareWithConfig(

			echoprometheus.MiddlewareConfig{
				// each 404 has own metric (not good)
				DoNotUseRequestPathFor404: true,
			},
		))
	}
}

func newHTTPErrorHandler(_ service.AppService) echo.HTTPErrorHandler {

	return func(err error, c echo.Context) {

		c.Echo().DefaultHTTPErrorHandler(err, c)

	}

}
func AssetsContentsMiddleware(e *echo.Echo, appService service.AppService, assetsFiles fs.FS) {

	Enabled := true
	if Enabled {

		e.StaticFS(consts.PathAuthAssets, assetsFiles)
		xlog.Info("Start serving embedded static content.")

	}
}

// func simpleAuthenticationMiddleware(container container.AppContainer) echo.MiddlewareFunc {
// 	return func(next echo.HandlerFunc) echo.HandlerFunc {
// 		return func(c echo.Context) error {
// 			if !hasAuthorization(c, container) {
// 				return c.JSON(http.StatusUnauthorized, false)
// 			}
// 			if err := next(c); err != nil {
// 				c.Error(err)
// 			}
// 			return nil
// 		}
// 	}
// }

// // hasAuthorization judges whether the user has the right to access the path.
// func hasAuthorization(c echo.Context, container container.AppContainer) bool {
// 	currentPath := c.Path()

// 	if pathPrefixed(currentPath, []string{consts.PathAuth}) {

// 		if pathPrefixed(currentPath, []string{".*/login$", ".*/logout$"}) {
// 			return true
// 		}

// 		accountDTO := container.Session().Account(c)
// 		if accountDTO == nil {
// 			return false
// 		}

// 		// TODO userDTO to identity

// 		role := "user"

// 		if role == "admin" || role == "user" {
// 			_ = container.Session().Save(c)
// 			return true
// 		}

// 		return false
// 	}
// 	return true
// }

// func pathPrefixed(cpath string, paths []string) bool {
// 	for i := range paths {

// 		if strings.HasPrefix(cpath, paths[i]) || regexp.MustCompile(paths[i]).Match([]byte(cpath)) {
// 			return true
// 		}
// 	}
// 	return false
// }