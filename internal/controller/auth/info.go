package auth

import (
	"go-auth/internal/config"
	controller "go-auth/internal/controller"

	"go-auth/internal/i18n"
	"go-auth/internal/service"
	"net/http"

	"github.com/labstack/echo/v4"

	xweb "go-auth/internal/web"
)

type AccountInfoController struct {
	appService service.AppService

	appConfig *config.AppConfig
	userLang  i18n.UserLang

	IsGET  bool
	IsPOST bool

	webCtxt echo.Context // webCtxt

	isAPIMode bool

	dto *InfoDTO
}

func (x *AccountInfoController) Info() error {
	// TODO sign out force

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

	// in any case redirect (or home hage)

}
func (x *AccountInfoController) responseDtoAsAPI() (err error) {

	dto := x.dto

	c := x.webCtxt

	controller.CsrfToHeader(c)
	return c.JSON(http.StatusOK, dto)

}
func (x *AccountInfoController) handleDto() (err error) {

	dto := x.dto
	c := x.webCtxt

	dto.IsAuth = xweb.IsSignedIn(c)

	return nil

}
func (x *AccountInfoController) responseDto() (err error) {

	return x.responseDtoAsAPI()

}

// NewAccountController is constructor.
func NewAccountInfoController(appService service.AppService, c echo.Context, isAPIMode bool) *AccountInfoController {
	appConfig := appService.Config()

	return &AccountInfoController{

		appService: appService,
		isAPIMode:  isAPIMode,
		appConfig:  appConfig,
		userLang:   controller.UserLang(c, appService),
		IsGET:      controller.IsGET(c),
		IsPOST:     controller.IsPOST(c),
		webCtxt:    c,
	}
}

type InfoDTO struct {
	IsAuth bool `json:"is_auth,omitempty"`
}

func (x *AccountInfoController) createDto() error {

	x.dto = &InfoDTO{}
	//

	return nil
}
