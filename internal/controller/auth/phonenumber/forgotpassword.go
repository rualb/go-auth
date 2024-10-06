package phonenumber

import (
	"go-auth/internal/config"
	"go-auth/internal/config/consts"
	controller "go-auth/internal/controller"

	"go-auth/internal/i18n"
	"go-auth/internal/mvc"
	"go-auth/internal/service"
	"go-auth/internal/tool/toolratelimit"
	"go-auth/internal/tool/toolstring"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

type AccountForgotPasswordController struct {
	appService service.AppService

	appConfig *config.AppConfig
	userLang  i18n.UserLang

	IsGET  bool
	IsPOST bool

	webCtxt echo.Context // webCtxt

	isAPIMode bool

	dto *ForgotPasswordDTO
}

func (x *AccountForgotPasswordController) Handler() error {

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

// NewAccountController is constructor.
func NewAccountForgotPasswordController(appService service.AppService, c echo.Context, isAPIMode bool) *AccountForgotPasswordController {

	appConfig := appService.Config()

	return &AccountForgotPasswordController{

		appService: appService,
		isAPIMode:  isAPIMode,
		appConfig:  appConfig,
		userLang:   controller.UserLang(c, appService),
		IsGET:      controller.IsGET(c),
		IsPOST:     controller.IsPOST(c),
		webCtxt:    c,
	}
}

type ForgotPasswordFormDTO struct {
	PhoneNumber string `form:"phone_number" json:"phone_number,omitempty"`
	NewPassword string `form:"new_password" json:"new_password,omitempty"`

	SecretCode   string `form:"secret_code" json:"secret_code,omitempty"`
	StepName     string `form:"step_name" json:"step_name,omitempty"`
	SecretString string `form:"secret_string" json:"secret_string,omitempty"`
}
type ForgotPasswordDTO struct {
	mvc.ModelBaseDTO
	ForgotPasswordFormDTO

	ReturnURL string `json:"return_url,omitempty"` // query:"return_url" form:"return_url"

	////////////////////////////
	IsSuccess            bool   `json:"is_success,omitempty"`
	IsFragment           bool   `json:"-"`
	IsRenderModeMessage  bool   `json:"is_render_mode_message,omitempty"`
	StatusMessage        string `json:"status_message,omitempty"`
	IsStatusMessageError bool   `json:"is_status_message_error,omitempty"`
	////////////////////////////
	NoRender bool `json:"-"`
	IsGET    bool `json:"-"`
	IsPOST   bool `json:"-"`
}

func (x *AccountForgotPasswordController) validateFields() {
	lang := x.userLang
	dto := x.dto

	{
		dto.PhoneNumber = toolstring.NormalizePhoneNumber(dto.PhoneNumber)
		dto.NewPassword = strings.TrimSpace(dto.NewPassword)

	}

	{
		v := dto.NewModelValidatorStr(lang, "phone_number", "Phone number" /*Lang*/, dto.PhoneNumber, consts.DefaultTextSize)
		v.Required()
		v.PhoneNumber()
		v.LengthRange(consts.PhoneNumberMinLength, consts.PhoneNumberMaxLength)

	}

	{
		v := dto.NewModelValidatorStr(lang, "new_password", "New password" /*Lang*/, dto.NewPassword, consts.DefaultTextSize)
		v.Required()
		v.Password(consts.PasswordMinLength)

	}

	{
		v := dto.NewModelValidatorStr(lang, "secret_code", "Secret code" /*Lang*/, dto.SecretCode, consts.DefaultTextSize)
		v.Required()
		v.LengthRange(consts.SecretCodeLength, consts.SecretCodeLength)

	}

	{

		if dto.IsStepID() {
			dto.RemoveModelError("secret_code")
			dto.RemoveModelError("new_password")
		}

		if dto.IsStepSecretCode() {
			// dto.RemoveModelError("secret_code")
			dto.RemoveModelError("new_password")
		}

		if dto.IsStepNewPassword() {
			dto.RemoveModelError("secret_code")
		}
	}
}

func (x *ForgotPasswordDTO) StepNext() {
	if x.IsStepID() {
		x.StepName = "secret_code"
	} else if x.IsStepSecretCode() {
		x.StepName = "new_password"
	}
}

// IsStepID returns true if StepName is empty.
func (x *ForgotPasswordDTO) IsStepID() bool {
	return x.StepName == ""
}

// IsStepSecretCode returns true if StepName is "SecretCode".
func (x *ForgotPasswordDTO) IsStepSecretCode() bool {
	return x.StepName == "secret_code"
}

// IsStepNewPassword returns true if StepName is "NewPassword".
func (x *ForgotPasswordDTO) IsStepNewPassword() bool {
	return x.StepName == "new_password"
}

func (x *AccountForgotPasswordController) createDto() error {

	x.dto = &ForgotPasswordDTO{}
	//
	dto := x.dto
	c := x.webCtxt

	if x.IsGET {

	} else if x.IsPOST {

		if err := c.Bind(&dto.ForgotPasswordFormDTO); err != nil {
			return err
		}

		{

			x.validateFields() // basic validation after UnMarshal
		}
	}

	if dto.PhoneNumber == "" {
		dto.PhoneNumber = x.appConfig.Identity.PhoneNumberPrefix
	}

	return nil
}

func (x *AccountForgotPasswordController) handleDto() error {

	dto := x.dto

	userLang := x.userLang
	c := x.webCtxt

	accountService := x.appService.AccountService()

	signInService := controller.SignInService(c, x.appService)

	if x.IsPOST {

		toolratelimit.RateLimitHuman()

		var user *service.UserAccount
		var err error

		sendSms := false
		gotoNextStep := false
		userExists := false
		userCanSignIn := false /*Sign up*/

		isInputValid := dto.IsModelValid()

		if isInputValid {

			user, err = accountService.FindByPhoneNumber(dto.PhoneNumber)

			if err != nil {
				return err // error e.g. db connection problem
			}

			if user == nil {
				dto.AddModelError("", userLang.Lang("No user found." /*Lang*/))
			} else {
				userExists = true
				userCanSignIn = signInService.CanSignIn(user) // no user with this PhoneNumber

				if !userCanSignIn {
					dto.AddModelError("", userLang.Lang("User account locked out." /*Lang*/))
				}
			}

		}

		if userExists && userCanSignIn { /*Sign up*/

			switch {
			case dto.IsStepID():
				{
					gotoNextStep = true
					sendSms = true
				}
			case dto.IsStepSecretCode():
				{

					isSecretCodeValid, err := accountService.ValidatePasscodeConfirmPhoneNumber(dto.SecretCode, dto.PhoneNumber, user)

					if err != nil {
						return err // error e.g. vault connect problem
					}
					if isSecretCodeValid {
						{
							secretString, err := accountService.GenerateTokenConfirmPhoneNumber(dto.PhoneNumber, user)
							dto.SecretString = secretString

							if err != nil {
								return err // error e.g. vault connect problem
							}

							gotoNextStep = true
						}
					} else {
						dto.IsStatusMessageError = true
						dto.StatusMessage = userLang.Lang("Secret code validation failed." /*Lang*/)
					}

				}
			case dto.IsStepNewPassword():
				{

					resetPassword, err := accountService.ValidateTokenConfirmPhoneNumber(dto.SecretString, dto.PhoneNumber, user)

					if err != nil {
						return err // error e.g. vault connect problem
					}

					if resetPassword {

						// set new password
						err = user.SetPassword(dto.NewPassword)
						if err != nil {
							return err
						}
						err = accountService.UpdateUserAccount(user)

						if err != nil {
							return err // error e.g. db connection problem
						}

						dto.IsSuccess = true
						dto.IsRenderModeMessage = true
						dto.StatusMessage = userLang.Lang("Your password has been changed." /*Lang*/)

						if dto.ReturnURL != "" {
							dto.NoRender = true
							err = c.Redirect(http.StatusFound /*302*/, dto.ReturnURL)
							if err != nil {
								return err //
							}
						}

						// TODO sign in force
						// TODO return goto ReturnURL

					}
				}
			}
		}

		if sendSms {
			toolratelimit.RateLimitMessage()

			secretCode, err := accountService.GeneratePasscodeConfirmPhoneNumber(dto.PhoneNumber, user)

			if err != nil {
				return err // error e.g. vault connect problem
			}

			x.appService.Messenger().SendSecretCodeToPhoneNumber(secretCode, dto.PhoneNumber, userLang.LangCode())
		}

		if gotoNextStep {
			dto.StepNext()
		}

	}

	return nil
}
func (x *AccountForgotPasswordController) responseDtoAsAPI() (err error) {

	dto := x.dto

	c := x.webCtxt

	controller.CsrfToHeader(c)
	return c.JSON(http.StatusOK, dto)

}

func (x *AccountForgotPasswordController) responseDtoAsMvc() (err error) {

	dto := x.dto
	appConfig := x.appConfig
	userLang := x.userLang
	c := x.webCtxt

	if dto.NoRender {
		return nil
	}

	data, err := mvc.NewModelWrap(c, dto, dto.IsFragment, "Password changing" /*Lang*/, appConfig, userLang)
	if err != nil {
		return err
	}
	err = c.Render(http.StatusOK, "forgot-password-phone-number.html", data)
	if err != nil {
		return err
	}

	return nil
}
func (x *AccountForgotPasswordController) responseDto() (err error) {

	if x.isAPIMode {
		return x.responseDtoAsAPI()
	} else {
		return x.responseDtoAsMvc()
	}
}
