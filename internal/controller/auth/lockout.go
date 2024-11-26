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

type LockoutController struct {
	appService service.AppService
	appConfig  *config.AppConfig
	userLang   i18n.UserLang

	IsGET  bool
	IsPOST bool

	webCtxt echo.Context // webCtxt

	isAPIMode bool

	dto *LockoutDTO
}

func (x *LockoutController) Handler() error {

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

func NewLockoutController(appService service.AppService, c echo.Context) *LockoutController {

	return &LockoutController{
		appService: appService,

		appConfig: appService.Config(),
		userLang:  controller.UserLang(c, appService),
		IsGET:     controller.IsGET(c),
		IsPOST:    controller.IsPOST(c),
		webCtxt:   c,
	}
}

type LockoutFormDTO struct {
}
type LockoutDTO struct {
	mvc.ModelBaseDTO
	LockoutFormDTO
	IsFragment bool `json:"-"`
	////////////////////////////
	NoRender bool `json:"-"`
	IsGET    bool `json:"-"`
	IsPOST   bool `json:"-"`
}

func (x *LockoutController) validateFields() {

}

func (x *LockoutController) createDTO() error {

	x.dto = &LockoutDTO{}
	//
	dto := x.dto
	c := x.webCtxt

	if x.IsGET {
		if err := c.Bind(&dto.LockoutFormDTO); err != nil {
			return err
		}

		{

			x.validateFields() // basic validation after UnMarshal
		}
	}

	return nil
}

func (x *LockoutController) handleDTO() error {

	return nil
}
func (x *LockoutController) responseDTOAsAPI() (err error) {

	return nil
}

func (x *LockoutController) responseDTOAsMvc() (err error) {

	dto := x.dto
	appConfig := x.appConfig
	lang := x.userLang
	c := x.webCtxt
	if x.isAPIMode {
		controller.CsrfToHeader(c)
		return c.JSON(http.StatusOK, dto)

	}

	data, err := mvc.NewModelWrap(c, dto, dto.IsFragment, "Lockout" /*Lang*/, appConfig, lang)
	if err != nil {
		return err
	}
	err = c.Render(http.StatusOK, "lockout.html", data)
	if err != nil {
		return err
	}

	return nil
}
func (x *LockoutController) responseDTO() (err error) {
	if x.isAPIMode {
		return x.responseDTOAsAPI()
	} else {
		return x.responseDTOAsMvc()
	}
}
