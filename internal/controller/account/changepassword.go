package account

import (
	"go-auth/internal/config"
	"go-auth/internal/config/consts"

	"go-auth/internal/controller"
	"go-auth/internal/i18n"
	"go-auth/internal/mvc"
	"go-auth/internal/service"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

type ChangePasswordController struct {
	appService service.AppService

	appConfig *config.AppConfig
	userLang  i18n.UserLang

	IsGET  bool
	IsPOST bool

	webCtxt echo.Context // webCtxt

	userAccount *service.UserAccount

	dto    *ChangePasswordDTO
	status int
}

func (x *ChangePasswordController) Handler() error {

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

// NewAccountController is constructor.
func NewChangePasswordController(appService service.AppService, c echo.Context) *ChangePasswordController {

	appConfig := appService.Config()

	return &ChangePasswordController{

		appService: appService,

		appConfig:   appConfig,
		userLang:    controller.UserLang(c, appService),
		IsGET:       controller.IsGET(c),
		IsPOST:      controller.IsPOST(c),
		userAccount: controller.GetAccountWithService(c, appService),
		webCtxt:     c,
	}
}

type ChangePasswordFormDTO struct {
	CurrentPassword string `form:"current_password" json:"current_password,omitempty"`
	NewPassword     string `form:"new_password" json:"new_password,omitempty"`
}
type ChangePasswordDTO struct {
	mvc.ModelBaseDTO
	ChangePasswordFormDTO

	NextURL string `json:"next,omitempty"` // query:"next" form:"next"

	////////////////////////////
	Status string `json:"status,omitempty"`
	// IsSuccess            bool   `json:"is_success,omitempty"`
	IsFragment bool `json:"-"`
	// IsRenderModeMessage  bool   `json:"is_render_mode_message,omitempty"`
	// StatusMessage string `json:"status_message,omitempty"`
	Message string `json:"message,omitempty"`
	// IsStatusMessageError bool   `json:"is_status_message_error,omitempty"`
	////////////////////////////
	NoRender bool `json:"-"`
	IsGET    bool `json:"-"`
	IsPOST   bool `json:"-"`
}

func (x *ChangePasswordController) validateFields() {
	lang := x.userLang
	dto := x.dto

	{

		dto.CurrentPassword = strings.TrimSpace(dto.CurrentPassword)
		dto.NewPassword = strings.TrimSpace(dto.NewPassword)

	}

	{
		v := dto.NewModelValidatorStr(lang, "current_password", "Current password" /*Lang*/, dto.CurrentPassword, consts.PasswordMaxLength)
		v.Required()
		// v.Password(consts.PasswordMinLength)

	}
	{
		v := dto.NewModelValidatorStr(lang, "new_password", "New password" /*Lang*/, dto.NewPassword, consts.PasswordMaxLength)
		v.Required()
		v.Password(consts.PasswordMinLength)

	}
}

func (x *ChangePasswordController) createDTO() error {

	x.dto = &ChangePasswordDTO{}
	//

	dto := x.dto
	c := x.webCtxt

	// fix binding problem (POST,GET,query)
	dto.NextURL = c.QueryParam("next")

	if x.IsGET {

	} else if x.IsPOST {

		if err := c.Bind(&dto.ChangePasswordFormDTO); err != nil {
			return err
		}

		{

			x.validateFields() // basic validation after UnMarshal
		}
	}

	return nil
}

func (x *ChangePasswordController) handleDTO() error {

	dto := x.dto
	userLang := x.userLang
	c := x.webCtxt

	botLimit := x.appService.Bot()

	if botLimit.LimitIPActivity(c.RealIP()) {
		x.status = http.StatusTooManyRequests
		return nil
	}

	accountService := x.appService.Account()

	if x.IsPOST {

		isInputValid := dto.IsModelValid()

		if isInputValid {

			var err error

			user := x.userAccount
			if user == nil {
				dto.AddError("", userLang.Lang("No user found." /*Lang*/))
				return nil
			}

			if botLimit.LimitAccountAccess(user.ID) {
				x.status = http.StatusTooManyRequests
				return nil
			}

			if user.CompareHashAndPassword(dto.CurrentPassword) {
				// set new password
				err = user.SetPassword(dto.NewPassword)
				if err != nil {
					return err
				}
				err = accountService.UpdateUserAccount(user)

				if err != nil {
					return err // error e.g. db connection problem
				}

				dto.NewPassword = ""
				dto.CurrentPassword = ""
				dto.Status = "success"
				// dto.IsSuccess = true
				// dto.IsRenderModeMessage = true
				dto.Message /*StatusMessage*/ = userLang.Lang("Your password has been changed." /*Lang*/)

			} else {
				dto.AddError("", userLang.Lang("Invalid password." /*Lang*/))
				return nil
			}

		}

	}

	return nil
}

func (x *ChangePasswordController) responseDTO() (err error) {

	c := x.webCtxt
	dto := x.dto

	if x.status == 0 {
		x.status = http.StatusOK
	}
	return c.JSON(x.status, dto)
}
