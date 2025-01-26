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

	dto    *SignoutDTO
	status int
}

func (x *AccountSignoutController) Signout() error {
	// TODO sign out force

	err := x.createDTO()
	if err != nil {
		return err
	}

	dto := x.dto
	userLang := x.userLang
	c := x.webCtxt

	if x.IsPOST {

		signInService := controller.SignInService(c, x.appService)

		signInService.SignOut()

		// dto.Password = ""
		dto.Status = "success"
		// dto.IsSuccess = true
		// dto.IsRenderModeMessage = true
		dto.Message /*StatusMessage*/ = userLang.Lang("You have successfully signed out of the application." /*Lang*/)

	}

	err = x.responseDTO()
	if err != nil {
		return err
	}

	return nil

	// in any case redirect (or home hage)

}
func (x *AccountSignoutController) responseDTOAsAPI() (err error) {

	dto := x.dto

	c := x.webCtxt

	if x.status == 0 {
		x.status = http.StatusOK
	}
	return c.JSON(x.status, dto)

}

func (x *AccountSignoutController) responseDTO() (err error) {

	return x.responseDTOAsAPI()

}

// NewAccountController is constructor.
func NewAccountSignoutController(appService service.AppService, c echo.Context) *AccountSignoutController {
	appConfig := appService.Config()

	return &AccountSignoutController{

		appService: appService,

		appConfig: appConfig,
		userLang:  controller.UserLang(c, appService),
		IsGET:     controller.IsGET(c),
		IsPOST:    controller.IsPOST(c),
		webCtxt:   c,
	}
}

type SignoutFormDTO struct {
}
type SignoutDTO struct {
	mvc.ModelBaseDTO
	SignoutFormDTO
	NextURL string `json:"next,omitempty"` // query:"next" form:"next"

	IsFragment bool `json:"-"`

	Status string `json:"status,omitempty"`
	// IsSuccess            bool   `json:"is_success,omitempty"`
	// IsRenderModeMessage  bool   `json:"is_render_mode_message,omitempty"`
	// StatusMessage string `json:"status_message,omitempty"`
	Message string `json:"message,omitempty"`
	// IsStatusMessageError bool   `json:"is_status_message_error,omitempty"`

	////////////////////////////
	NoRender bool `json:"-"`
	IsGET    bool `json:"-"`
	IsPOST   bool `json:"-"`
}

func (x *AccountSignoutController) validateFields() {

}

func (x *AccountSignoutController) createDTO() error {

	x.dto = &SignoutDTO{}
	//
	dto := x.dto
	c := x.webCtxt

	// fix binding problem (POST,GET,query)
	dto.NextURL = c.QueryParam("next")

	// if err := c.Bind(dto); err != nil {
	// 	return nil, err
	// }

	{

		x.validateFields() // basic validation after UnMarshal
	}

	return nil
}
