package email

import (
	"go-auth/internal/config"
	"go-auth/internal/config/consts"
	controller "go-auth/internal/controller"
	"strings"

	"go-auth/internal/i18n"
	"go-auth/internal/mvc"
	"go-auth/internal/service"
	"go-auth/internal/util/utilratelimit"
	"net/http"

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
	Email       string `form:"email" json:"email,omitempty"`
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
		dto.Email = strings.TrimSpace(dto.Email)

		dto.NewPassword = strings.TrimSpace(dto.NewPassword)

	}
	{
		v := dto.NewModelValidatorStr(lang, "email", "Email" /*Lang*/, dto.Email, consts.DefaultTextSize)
		v.Required()
		v.Email(consts.EmailMinLength)

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
			dto.RemoveError("secret_code")
			dto.RemoveError("new_password")
		}

		if dto.IsStepSecretCode() {
			// dto.RemoveError("secret_code")
			dto.RemoveError("new_password")
		}

		if dto.IsStepNewPassword() {
			dto.RemoveError("secret_code")
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

func (x *AccountForgotPasswordController) createDTO() error {

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

	return nil
}

func (x *AccountForgotPasswordController) handleDTO() error {

	dto := x.dto

	userLang := x.userLang
	c := x.webCtxt

	accountService := x.appService.AccountService()

	signInService := controller.SignInService(c, x.appService)

	if x.IsPOST {

		utilratelimit.RateLimitHuman()

		var user *service.UserAccount
		var err error

		sendSms := false
		gotoNextStep := false
		userExists := false
		userCanSignIn := false /*Sign up*/

		Email := dto.Email

		isInputValid := dto.IsModelValid()

		if isInputValid {

			user, err = accountService.FindByNormalizedEmail(Email)

			if err != nil {
				return err // error e.g. db connection problem
			}

			if user == nil {
				dto.AddError("", userLang.Lang("No user found." /*Lang*/))
			} else {
				userExists = true
				userCanSignIn = signInService.CanSignIn(user) // no user with this Email

				if !userCanSignIn {
					dto.AddError("", userLang.Lang("User account locked out." /*Lang*/))
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

					isSecretCodeValid, err := accountService.ValidatePasscodeConfirmEmail(dto.SecretCode, Email, user)

					if err != nil {
						return err // error e.g. vault connect problem
					}
					if isSecretCodeValid {
						{
							secretString, err := accountService.GenerateTokenConfirmEmail(Email, user)
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

					resetPassword, err := accountService.ValidateTokenConfirmEmail(dto.SecretString, Email, user)

					if err != nil {
						return err // error e.g. vault connect problem
					}

					if resetPassword {

						// set new password
						err := user.SetPassword(dto.NewPassword)
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
								return err
							}
						}

						// TODO sign in force
						// TODO return goto ReturnURL

					}
				}
			}
		}

		if sendSms {

			utilratelimit.RateLimitMessage()

			secretCode, err := accountService.GeneratePasscodeConfirmEmail(Email, user)

			if err != nil {
				return err // error e.g. vault connect problem
			}

			x.appService.Messenger().SendSecretCodeToEmail(secretCode, Email, userLang.LangCode())
		}

		if gotoNextStep {
			dto.StepNext()
		}

	}

	return nil
}
func (x *AccountForgotPasswordController) responseDTOAsAPI() (err error) {

	dto := x.dto

	c := x.webCtxt

	controller.CsrfToHeader(c)
	return c.JSON(http.StatusOK, dto)

}

func (x *AccountForgotPasswordController) responseDTOAsMvc() (err error) {

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
	err = c.Render(http.StatusOK, "forgot-password-email.html", data)
	if err != nil {
		return err
	}

	return nil
}
func (x *AccountForgotPasswordController) responseDTO() (err error) {

	if x.isAPIMode {
		return x.responseDTOAsAPI()
	} else {
		return x.responseDTOAsMvc()
	}
}
