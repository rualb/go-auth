package auth

import (
	"go-auth/internal/config"
	controller "go-auth/internal/controller"
	"go-auth/internal/service"

	"go-auth/internal/i18n"
	"go-auth/internal/mvc"
	"net/http"

	"github.com/labstack/echo/v4"
)

type AccessDeniedController struct {
	appService service.AppService
	appConfig  *config.AppConfig
	userLang   i18n.UserLang

	IsGET  bool
	IsPOST bool

	webCtxt echo.Context // webCtxt

	isAPIMode bool

	dto *AccessDeniedDTO
}

func (x *AccessDeniedController) Handler() error {

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

func NewAccessDeniedController(appService service.AppService, c echo.Context) *AccessDeniedController {

	return &AccessDeniedController{
		appService: appService,

		appConfig: appService.Config(),
		userLang:  controller.UserLang(c, appService),
		IsGET:     controller.IsGET(c),
		IsPOST:    controller.IsPOST(c),
		webCtxt:   c,
	}
}

type AccessDeniedFormDTO struct {
}
type AccessDeniedDTO struct {
	mvc.ModelBaseDTO
	AccessDeniedFormDTO
	IsFragment bool `json:"-"`
	////////////////////////////
	NoRender bool `json:"-"`
}

func (x *AccessDeniedController) validateFields() {

}

func (x *AccessDeniedController) createDto() error {

	x.dto = &AccessDeniedDTO{}
	//
	dto := x.dto
	c := x.webCtxt

	if x.IsGET {
		if err := c.Bind(&dto.AccessDeniedFormDTO); err != nil {
			return err
		}

		{

			x.validateFields() // basic validation after UnMarshal
		}
	}

	return nil
}

func (x *AccessDeniedController) handleDto() error {

	return nil
}

func (x *AccessDeniedController) responseDtoAsAPI() (err error) {

	dto := x.dto

	c := x.webCtxt

	controller.CsrfToHeader(c)
	return c.JSON(http.StatusOK, dto)

}

func (x *AccessDeniedController) responseDtoAsMvc() (err error) {

	dto := x.dto
	appConfig := x.appConfig
	lang := x.userLang
	c := x.webCtxt

	data, err := mvc.NewModelWrap(c, dto, dto.IsFragment, "Access denied" /*Lang*/, appConfig, lang)
	if err != nil {
		return err
	}
	err = c.Render(http.StatusOK, "access-denied.html", data)
	if err != nil {
		return err
	}

	return nil
}

func (x *AccessDeniedController) responseDto() (err error) {
	if x.isAPIMode {
		return x.responseDtoAsAPI()
	} else {
		return x.responseDtoAsMvc()
	}
}
