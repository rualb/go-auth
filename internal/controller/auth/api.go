package auth

import (
	"go-auth/internal/service"
	"net/http"

	"github.com/labstack/echo/v4"

	xweb "go-auth/internal/web"
)

//////////////////////////////////////////////////

type StatusDTO struct {
	Input struct{}
	Meta  struct {
		Status int
	}
	Output struct {
		IsAuth bool `json:"is_auth,omitempty"`
	}
}

type StatusAPIController struct {
	appService service.AppService
	webCtxt    echo.Context // webCtxt
}

// NewAccountController is constructor.
func NewStatusAPIController(appService service.AppService, c echo.Context) *StatusAPIController {
	return &StatusAPIController{
		appService: appService,
		webCtxt:    c,
	}
}
func (x *StatusAPIController) Handler() (err error) {

	dto := &StatusDTO{}
	//
	meta := &dto.Meta
	output := &dto.Output
	c := x.webCtxt
	//
	output.IsAuth = xweb.IsSignedIn(c)
	//
	// controller.CsrfToHeader(c)
	//
	if meta.Status == 0 {
		meta.Status = http.StatusOK
	}
	return c.JSON(meta.Status, output)

}

//////////////////////////////////////////////////

type ConfigDTO struct {
	Input struct{}
	Meta  struct {
		Status int
	}
	Output struct{}
}

type ConfigAPIController struct {
	appService service.AppService
	webCtxt    echo.Context // webCtxt
}

// NewConfigAPIController is constructor.
func NewConfigAPIController(appService service.AppService, c echo.Context) *ConfigAPIController {
	return &ConfigAPIController{
		appService: appService,
		webCtxt:    c,
	}
}

func (x *ConfigAPIController) Handler() (err error) {

	dto := &ConfigDTO{}
	//
	meta := &dto.Meta
	output := &dto.Output
	c := x.webCtxt
	//

	//
	if meta.Status == 0 {
		meta.Status = http.StatusOK
	}
	return c.JSON(meta.Status, output)

}

//////////////////////////////////////////////////
