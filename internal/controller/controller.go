package controller

import (
	"go-auth/internal/i18n"
	"go-auth/internal/service"

	xtoken "go-auth/internal/token"
	xweb "go-auth/internal/web"

	"github.com/labstack/echo/v4"
)

// func LangCode(c echo.Context) string {
//		lang, _ := c.Get("lang_code").(string)
//		return lang
// }

func UserLang(c echo.Context, appLang i18n.AppLang) i18n.UserLang {

	lang, _ := c.Get("lang_code").(string)
	return appLang.UserLang(lang)
}

func IsGET(c echo.Context) bool {
	return c.Request().Method == "GET"
}

func IsPOST(c echo.Context) bool {
	return c.Request().Method == "POST"
}

func CsrfToHeader(c echo.Context) {
	csrf, _ := c.Get("_csrf").(string)
	c.Response().Header().Set("X-CSRF-Token", csrf)
}

func newTokenPersist(c echo.Context, appService service.AppService) xtoken.TokenPersist {
	return xweb.NewTokenPersist(c, appService)
}
func SignInService(c echo.Context, appService service.AppService) service.SignInService {
	return appService.SignInService(newTokenPersist(c, appService))
}