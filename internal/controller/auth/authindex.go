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
	isAPIMode bool

	appService service.AppService
	appConfig  *config.AppConfig
	userLang   i18n.UserLang

	IsGET  bool
	IsPOST bool

	webCtxt echo.Context // webCtxt

	dto *AuthIndexDTO
}

func (x *IndexController) Handler() error {

	err := x.createDto()
	if err != nil {
		return err
	}

	err = x.handleDto()
	if err != nil {
		return err
	}

	err = x.responseDto()
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
		IsAuthPhone  bool   `json:"is_auth_phone,omitempty"`
		IsAuthEmail  bool   `json:"is_auth_email,omitempty"`
		IsAuthSignup bool   `json:"is_auth_signup,omitempty"`
		IsAuthForgot bool   `json:"is_auth_forgot,omitempty"`
	}
	Title    string
	LangData map[string]string
}

func (x *IndexController) validateFields() {

}

func (x *IndexController) createDto() error {

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

func (x *IndexController) handleDto() error {

	dto := x.dto
	// c := x.webCtxt

	userLang := x.userLang
	dto.LangCode = userLang.LangCode()
	dto.Title = userLang.Lang("Authentication") // TODO /*Lang*/

	cnf := &dto.AppConfig

	cnf.AppTitle = x.appConfig.Title
	cnf.TmTitle = fmt.Sprintf("© %v %s", time.Now().Year(), x.appConfig.Title)
	cnf.IsAuthPhone = x.appConfig.Identity.IsAuthPhone
	cnf.IsAuthEmail = x.appConfig.Identity.IsAuthEmail
	cnf.IsAuthSignup = x.appConfig.Identity.IsAuthSignup
	cnf.IsAuthForgot = x.appConfig.Identity.IsAuthForgot

	dto.LangData = userLang.LangData()
	return nil
}

func (x *IndexController) responseDtoAsAPI() (err error) {

	return nil
}

func (x *IndexController) responseDtoAsMvc() (err error) {

	dto := x.dto
	appConfig := x.appConfig
	lang := x.userLang
	c := x.webCtxt

	data, err := mvc.NewModelWrap(c, dto, dto.IsFragment, "Auth" /*Lang*/, appConfig, lang)
	if err != nil {
		return err
	}

	err = c.Render(http.StatusOK, "auth.html", data)

	if err != nil {
		return err
	}

	return nil
}
func (x *IndexController) responseDto() (err error) {
	if x.isAPIMode {
		return x.responseDtoAsAPI()
	} else {
		return x.responseDtoAsMvc()
	}
}
