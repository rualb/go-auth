package controller

import (
	"go-auth/internal/service"
	"net/http"

	"github.com/labstack/echo/v4"
)

type PingController interface {
	Ping(c echo.Context) error
}

type pingController struct {
}

func NewPingController(_ service.AppService) PingController {
	return &pingController{}
}

func (controller *pingController) Ping(c echo.Context) error {
	res := "pong"

	return c.String(http.StatusOK, res)
}
