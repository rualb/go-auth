package phonenumber

import (
	"go-auth/internal/config"
	"go-auth/internal/config/consts"
	controller "go-auth/internal/controller"
	"go-auth/internal/util/utilstring"

	"go-auth/internal/i18n"
	"go-auth/internal/mvc"
	"go-auth/internal/service"
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

	dto    *ForgotPasswordDTO
	status int
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
func NewAccountForgotPasswordController(appService service.AppService, c echo.Context) *AccountForgotPasswordController {

	appConfig := appService.Config()

	return &AccountForgotPasswordController{

		appService: appService,

		appConfig: appConfig,
		userLang:  controller.UserLang(c, appService),
		IsGET:     controller.IsGET(c),
		IsPOST:    controller.IsPOST(c),
		webCtxt:   c,
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
		dto.PhoneNumber = utilstring.NormalizePhoneNumber(dto.PhoneNumber)
		dto.NewPassword = strings.TrimSpace(dto.NewPassword)

	}

	{
		v := dto.NewModelValidatorStr(lang, "phone_number", "Phone number" /*Lang*/, dto.PhoneNumber, consts.DefaultTextLength)
		v.Required()
		v.PhoneNumber()
		v.LengthRange(consts.PhoneNumberMinLength, consts.PhoneNumberMaxLength)

	}

	{
		v := dto.NewModelValidatorStr(lang, "new_password", "New password" /*Lang*/, dto.NewPassword, consts.PasswordMaxLength)
		v.Required()
		v.Password(consts.PasswordMinLength)

	}

	{
		v := dto.NewModelValidatorStr(lang, "secret_code", "Secret code" /*Lang*/, dto.SecretCode, consts.DefaultTextLength)
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

	if dto.PhoneNumber == "" {
		dto.PhoneNumber = x.appConfig.Identity.PhoneNumberPrefix
	}

	return nil
}

func (x *AccountForgotPasswordController) handleDTO() error {

	dto := x.dto

	userLang := x.userLang
	c := x.webCtxt

	botLimit := x.appService.Bot()

	if botLimit.LimitIPActivity(c.RealIP()) {
		x.status = http.StatusTooManyRequests
		return nil
	}

	accountService := x.appService.Account()

	signInService := controller.SignInService(c, x.appService)

	if x.IsPOST {

		var user *service.UserAccount
		var err error

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
				dto.AddError("", userLang.Lang("No user found." /*Lang*/))
			} else {
				userExists = true
				userCanSignIn = signInService.CanSignIn(user) // no user with this PhoneNumber

				if !userCanSignIn {
					dto.AddError("", userLang.Lang("User account locked out." /*Lang*/))
				}
			}

		}

		if userExists && userCanSignIn { /*Sign up*/

			if botLimit.LimitAccountAccess(user.ID) {
				x.status = http.StatusTooManyRequests
				return nil
			}

			switch {
			case dto.IsStepID():
				{
					gotoNextStep = true
					sendSecretMsg := true

					if sendSecretMsg {
						if botLimit.LimitUserMessage(dto.PhoneNumber, user.ID) {
							x.status = http.StatusTooManyRequests
							return nil
						}

						secretCode, err := accountService.GeneratePasscodeConfirmPhoneNumber(dto.PhoneNumber, user)

						if err != nil {
							return err // error e.g. vault connect problem
						}

						x.appService.Messenger().SendSecretCodeToPhoneNumber(secretCode, dto.PhoneNumber, userLang.LangCode())
					}
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

		if gotoNextStep {
			dto.StepNext()
		}

	}

	return nil
}
func (x *AccountForgotPasswordController) responseDTOAsAPI() (err error) {

	dto := x.dto

	c := x.webCtxt

	if x.status == 0 {
		x.status = http.StatusOK
	}
	return c.JSON(x.status, dto)

}

func (x *AccountForgotPasswordController) responseDTO() (err error) {

	return x.responseDTOAsAPI()

}
