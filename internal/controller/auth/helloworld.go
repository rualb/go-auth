package auth

import (
	"go-auth/internal/config"
	controller "go-auth/internal/controller"
	"go-auth/internal/service"
	"time"

	"go-auth/internal/i18n"
	"go-auth/internal/mvc"
	"net/http"

	"github.com/labstack/echo/v4"
)

type HelloWorldController struct {
	isAPIMode bool

	appService service.AppService

	appConfig *config.AppConfig
	userLang  i18n.UserLang

	IsGET  bool
	IsPOST bool

	webCtxt echo.Context // webCtxt

	dto *HelloWorldDTO
}

func (x *HelloWorldController) Handler() error {

	err := x.createDto()
	if err != nil {
		return err
	}

	//
	err = x.handleDto()
	if err != nil {
		return err
	}

	return x.responseDto()

}

func NewHelloWorldController(appService service.AppService, c echo.Context) *HelloWorldController {

	return &HelloWorldController{
		appService: appService,

		appConfig: appService.Config(),
		userLang:  controller.UserLang(c, appService),
		IsGET:     controller.IsGET(c),
		IsPOST:    controller.IsPOST(c),
		webCtxt:   c,
	}
}

type HelloWorldDTO struct {
	mvc.ModelBaseDTO

	Name       string    `query:"name"`
	IsFragment bool      `query:"is_fragment"`
	Time       time.Time `query:"time"` // 2006-01-02T15:04:05Z
	NowStr     string
}

func (x *HelloWorldController) validateFields() {

}

func (x *HelloWorldController) createDto() error {

	x.dto = &HelloWorldDTO{}
	//

	dto := x.dto

	c := x.webCtxt

	if x.IsGET {
		if err := c.Bind(dto); err != nil {
			return err
		}

		{
			x.validateFields() // basic validation after UnMarshal
		}
	}

	return nil
}

func (x *HelloWorldController) handleDto() error {

	dto := x.dto

	userLang := x.userLang

	dto.AddModelError("", userLang.Lang("HelloWorld" /*Lang*/))
	dto.NowStr = time.Now().Truncate(time.Second).Format(time.RFC3339)

	return nil
}
func (x *HelloWorldController) responseDtoAsAPI() (err error) {

	return nil
}

func (x *HelloWorldController) responseDtoAsMvc() (err error) {

	dto := x.dto
	appConfig := x.appConfig
	userLang := x.userLang
	c := x.webCtxt

	data, err := mvc.NewModelWrap(c, dto, dto.IsFragment, "Hello world" /*Lang*/, appConfig, userLang)
	if err != nil {
		return err
	}
	err = c.Render(http.StatusOK, "hello-world.html", data)
	if err != nil {
		return err
	}

	return nil
}
func (x *HelloWorldController) responseDto() (err error) {

	if x.isAPIMode {
		return x.responseDtoAsAPI()
	} else {
		return x.responseDtoAsMvc()
	}
}
