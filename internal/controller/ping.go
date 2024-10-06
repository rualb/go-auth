package controller

import (
	"go-auth/internal/service"
	"net/http"
	"strconv"
	"strings"

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

	if name := c.QueryParam("name"); name != "" {
		res = res + " " + name
	}

	if length := c.QueryParam("length"); length != "" {

		lengthI, _ := strconv.Atoi(length)
		lengthI = min(max(lengthI, 0), 32000)

		res = res + " " + strings.Repeat("A", lengthI)
	}

	return c.String(http.StatusOK, res)
}
