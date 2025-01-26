package email

import (
	"go-auth/internal/config"
	"go-auth/internal/config/consts"
	controller "go-auth/internal/controller"
	"strings"

	"go-auth/internal/i18n"
	"go-auth/internal/mvc"
	"go-auth/internal/service"
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
	Email       string `form:"email" json:"email,omitempty"`
	NewPassword string `form:"new_password" json:"new_password,omitempty"`

	Passcode string `form:"passcode" json:"passcode,omitempty"`
	Step     string `form:"step" json:"step,omitempty"`
	Token    string `form:"token" json:"token,omitempty"`
}
type ForgotPasswordDTO struct {
	mvc.ModelBaseDTO
	ForgotPasswordFormDTO

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

func (x *AccountForgotPasswordController) validateFields() {
	lang := x.userLang
	dto := x.dto

	{
		dto.Email = strings.TrimSpace(dto.Email)

		dto.NewPassword = strings.TrimSpace(dto.NewPassword)

		dto.Passcode = strings.TrimSpace(dto.Passcode)

	}
	{
		v := dto.NewModelValidatorStr(lang, "email", "Email" /*Lang*/, dto.Email, consts.DefaultTextLength)
		v.Required()
		v.Email(consts.EmailMinLength)

	}

	{
		v := dto.NewModelValidatorStr(lang, "new_password", "New password" /*Lang*/, dto.NewPassword, consts.PasswordMaxLength)
		v.Required()
		v.Password(consts.PasswordMinLength)

	}

	{
		v := dto.NewModelValidatorStr(lang, "passcode", "Secret code" /*Lang*/, dto.Passcode, consts.DefaultTextLength)
		v.Required()
		v.Digits()
		v.LengthRange(consts.PasscodeLength, consts.PasscodeLength)

	}

	{

		if dto.IsStepID() {
			dto.RemoveError("passcode")
			dto.RemoveError("new_password")
		}

		if dto.IsStepPasscode() {
			// dto.RemoveError("passcode")
			dto.RemoveError("new_password")
		}

		if dto.IsStepNewPassword() {
			dto.RemoveError("passcode")
		}
	}
}

func (x *ForgotPasswordDTO) StepNext() {
	if x.IsStepID() {
		x.Step = "passcode"
	} else if x.IsStepPasscode() {
		x.Step = "new_password"
	}
}

// IsStepID returns true if Step is empty.
func (x *ForgotPasswordDTO) IsStepID() bool {
	return x.Step == ""
}

// IsStepPasscode returns true if Step is "passcode".
func (x *ForgotPasswordDTO) IsStepPasscode() bool {
	return x.Step == "passcode"
}

// IsStepNewPassword returns true if Step is "new_password".
func (x *ForgotPasswordDTO) IsStepNewPassword() bool {
	return x.Step == "new_password"
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

						botLimit.LimitUserMessage(Email, user.ID)

						passcode, err := accountService.GeneratePasscodeConfirmEmail(Email, user)

						if err != nil {
							return err // error e.g. vault connect problem
						}

						x.appService.Messenger().SendPasscodeToEmail(passcode, Email, userLang.LangCode())
					}
				}
			case dto.IsStepPasscode():
				{

					isPasscodeValid, err := accountService.ValidatePasscodeConfirmEmail(dto.Passcode, Email, user)

					if err != nil {
						return err // error e.g. vault connect problem
					}
					if isPasscodeValid {
						{
							token, err := accountService.GenerateTokenConfirmEmail(Email, user)
							dto.Token = token

							if err != nil {
								return err // error e.g. vault connect problem
							}

							gotoNextStep = true
						}
					} else {
						// dto.IsStatusMessageError = true
						// dto.Status = ""
						dto.Message /*StatusMessage*/ = userLang.Lang("Secret code validation failed." /*Lang*/)
					}

				}
			case dto.IsStepNewPassword():
				{

					resetPassword, err := accountService.ValidateTokenConfirmEmail(dto.Token, Email, user)

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

						dto.NewPassword = ""
						dto.Status = "success"
						// dto.IsSuccess = true
						// dto.IsRenderModeMessage = true
						dto.Message /*StatusMessage*/ = userLang.Lang("Your password has been changed." /*Lang*/)

						if dto.NextURL != "" {
							dto.NoRender = true
							err = c.Redirect(http.StatusFound /*302*/, dto.NextURL)
							if err != nil {
								return err
							}
						}

						// TODO sign in force
						// TODO return goto NextURL

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
