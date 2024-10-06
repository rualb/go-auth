package auth

import (
	"go-auth/internal/config"
	controller "go-auth/internal/controller"
	"go-auth/internal/mvc"

	"go-auth/internal/i18n"
	"go-auth/internal/service"
	"net/http"

	"github.com/labstack/echo/v4"
)

type AccountSignoutController struct {
	appService service.AppService

	appConfig *config.AppConfig
	userLang  i18n.UserLang

	IsGET  bool
	IsPOST bool

	webCtxt echo.Context // webCtxt

	isAPIMode bool

	dto *SignoutDTO
}

func (x *AccountSignoutController) Signout() error {
	// TODO sign out force

	err := x.createDto()
	if err != nil {
		return err
	}

	dto := x.dto
	userLang := x.userLang
	c := x.webCtxt

	if x.IsPOST {

		signInService := controller.SignInService(c, x.appService)

		signInService.SignOut()

		dto.IsSuccess = true
		dto.IsRenderModeMessage = true
		dto.StatusMessage = userLang.Lang("You have successfully signed out of the application." /*Lang*/)

	}

	err = x.responseDto()
	if err != nil {
		return err
	}

	return nil

	// in any case redirect (or home hage)

}
func (x *AccountSignoutController) responseDtoAsAPI() (err error) {

	dto := x.dto

	c := x.webCtxt

	controller.CsrfToHeader(c)
	return c.JSON(http.StatusOK, dto)

}

func (x *AccountSignoutController) responseDtoAsMvc() (err error) {

	dto := x.dto
	appConfig := x.appConfig
	lang := x.userLang
	c := x.webCtxt

	data, err := mvc.NewModelWrap(c, dto, dto.IsFragment, "Sign out" /*Lang*/, appConfig, lang)
	if err != nil {
		return err
	}
	err = c.Render(http.StatusOK, "signout.html", data)
	if err != nil {
		return err
	}

	return nil
}
func (x *AccountSignoutController) responseDto() (err error) {
	if x.isAPIMode {
		return x.responseDtoAsAPI()
	} else {
		return x.responseDtoAsMvc()
	}
}

// NewAccountController is constructor.
func NewAccountSignoutController(appService service.AppService, c echo.Context, isAPIMode bool) *AccountSignoutController {
	appConfig := appService.Config()

	return &AccountSignoutController{

		appService: appService,
		isAPIMode:  isAPIMode,
		appConfig:  appConfig,
		userLang:   controller.UserLang(c, appService),
		IsGET:      controller.IsGET(c),
		IsPOST:     controller.IsPOST(c),
		webCtxt:    c,
	}
}

type SignoutFormDTO struct {
}
type SignoutDTO struct {
	mvc.ModelBaseDTO
	SignoutFormDTO
	ReturnURL string `json:"return_url,omitempty"` // query:"return_url" form:"return_url"

	IsFragment bool `json:"-"`

	IsSuccess            bool   `json:"is_success,omitempty"`
	IsRenderModeMessage  bool   `json:"is_render_mode_message,omitempty"`
	StatusMessage        string `json:"status_message,omitempty"`
	IsStatusMessageError bool   `json:"is_status_message_error,omitempty"`

	////////////////////////////
	NoRender bool `json:"-"`
	IsGET    bool `json:"-"`
	IsPOST   bool `json:"-"`
}

func (x *AccountSignoutController) validateFields() {

}

func (x *AccountSignoutController) createDto() error {

	x.dto = &SignoutDTO{}
	//
	dto := x.dto
	c := x.webCtxt

	// fix binding problem (POST,GET,query)
	dto.ReturnURL = c.QueryParam("return_url")

	// if err := c.Bind(dto); err != nil {
	// 	return nil, err
	// }

	{

		x.validateFields() // basic validation after UnMarshal
	}

	return nil
}
