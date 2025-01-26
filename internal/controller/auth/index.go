package auth

import (
	"fmt"
	"go-auth/internal/config"
	controller "go-auth/internal/controller"
	"go-auth/internal/service"
	"time"

	"go-auth/internal/i18n"
	"go-auth/internal/mvc"
	"net/http"

	"github.com/labstack/echo/v4"
)

type IndexController struct {
	appService service.AppService
	appConfig  *config.AppConfig
	userLang   i18n.UserLang

	IsGET  bool
	IsPOST bool

	webCtxt echo.Context // webCtxt

	dto *AuthIndexDTO
}

func (x *IndexController) Handler() error {

	err := x.createDTO()
	if err != nil {
		return err
	}

	err = x.handleDTO()
	if err != nil {
		return err
	}

	err = x.responseDTO()
	if err != nil {
		return err
	}

	return nil
}

func NewIndexController(appService service.AppService, c echo.Context) *IndexController {

	return &IndexController{
		appService: appService,
		appConfig:  appService.Config(),
		userLang:   controller.UserLang(c, appService),
		IsGET:      controller.IsGET(c),
		IsPOST:     controller.IsPOST(c),
		webCtxt:    c,
	}
}

type AuthIndexFormDTO struct {
}

type AuthIndexDTO struct {
	mvc.ModelBaseDTO
	AuthIndexFormDTO
	IsFragment bool `json:"-"`
	////////////////////////////

	LangCode  string
	AppConfig struct {
		AppTitle     string `json:"app_title,omitempty"`
		TmTitle      string `json:"tm_title,omitempty"`
		IsAuthTel    bool   `json:"is_auth_tel,omitempty"`
		IsAuthEmail  bool   `json:"is_auth_email,omitempty"`
		IsAuthSignup bool   `json:"is_auth_signup,omitempty"`
		IsAuthForgot bool   `json:"is_auth_forgot,omitempty"`
	}
	Title     string
	LangWords map[string]string
}

func (x *IndexController) validateFields() {

}

func (x *IndexController) createDTO() error {

	x.dto = &AuthIndexDTO{}
	//

	dto := x.dto
	c := x.webCtxt

	if x.IsGET {
		if err := c.Bind(&dto.AuthIndexFormDTO); err != nil {
			return err
		}

		x.validateFields() // basic validation after UnMarshal
	}

	return nil
}

func (x *IndexController) handleDTO() error {

	dto := x.dto
	// c := x.webCtxt

	userLang := x.userLang
	dto.LangCode = userLang.LangCode()
	dto.Title = userLang.Lang("Authentication") // TODO /*Lang*/

	cfg := &dto.AppConfig

	cfg.AppTitle = x.appConfig.Title
	cfg.TmTitle = fmt.Sprintf("%s © %d", x.appConfig.Title, time.Now().Year())
	cfg.IsAuthTel = x.appConfig.Identity.IsAuthTel
	cfg.IsAuthEmail = x.appConfig.Identity.IsAuthEmail
	cfg.IsAuthSignup = x.appConfig.Identity.IsAuthSignup
	cfg.IsAuthForgot = x.appConfig.Identity.IsAuthForgot

	dto.LangWords = userLang.LangWords()
	return nil
}

func (x *IndexController) responseDTOAsMvc() (err error) {

	dto := x.dto
	appConfig := x.appConfig
	lang := x.userLang
	c := x.webCtxt

	data, err := mvc.NewModelWrap(c, dto, dto.IsFragment, "Auth" /*Lang*/, appConfig, lang)
	if err != nil {
		return err
	}

	err = c.Render(http.StatusOK, "index.html", data)

	if err != nil {
		return err
	}

	return nil
}
func (x *IndexController) responseDTO() (err error) {

	return x.responseDTOAsMvc()

}
