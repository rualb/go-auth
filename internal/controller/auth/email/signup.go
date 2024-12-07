package email

import (
	"fmt"
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

type AccountSignupController struct {
	appService service.AppService
	appConfig  *config.AppConfig
	userLang   i18n.UserLang

	IsGET  bool
	IsPOST bool

	webCtxt echo.Context // webCtxt

	dto    *SignupDTO
	status int
}

func (x *AccountSignupController) Handler() error {
	// TODO sign out force

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
func NewAccountSignupController(appService service.AppService, c echo.Context) *AccountSignupController {

	appConfig := appService.Config()

	return &AccountSignupController{

		appService: appService,

		appConfig: appConfig,
		userLang:  controller.UserLang(c, appService),
		IsGET:     controller.IsGET(c),
		IsPOST:    controller.IsPOST(c),
		webCtxt:   c,
	}
}

type SignupFormDTO struct {
	Email string `form:"email" json:"email,omitempty"`

	Password string `form:"password" json:"password,omitempty"`

	SecretCode   string `form:"secret_code" json:"secret_code,omitempty"`
	StepName     string `form:"step_name" json:"step_name,omitempty"`
	SecretString string `form:"secret_string" json:"secret_string,omitempty"`
}

type SignupDTO struct {
	mvc.ModelBaseDTO
	SignupFormDTO
	//
	// StateString  string `form:"state_string"` //=><input value="{{ .Model.StateString }}" name="StateString" type="hidden" />
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

func (x *AccountSignupController) validateFields() {

	lang := x.userLang
	dto := x.dto

	{
		dto.Email = strings.TrimSpace(dto.Email)
		dto.Password = strings.TrimSpace(dto.Password)
	}

	{
		v := dto.NewModelValidatorStr(lang, "email", "Email" /*Lang*/, dto.Email, consts.DefaultTextLength)
		v.Required()
		v.Email(consts.EmailMinLength)

	}

	{
		v := dto.NewModelValidatorStr(lang, "password", "Password" /*Lang*/, dto.Password, consts.PasswordMaxLength)
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
			dto.RemoveError("password")
		}

		if dto.IsStepSecretCode() {
			// dto.RemoveError("secret_code")
			dto.RemoveError("password")
		}

		if dto.IsStepPassword() {
			dto.RemoveError("secret_code")
		}
	}
}

func (x *SignupDTO) StepNext() {
	if x.IsStepID() {
		x.StepName = "secret_code"
	} else if x.IsStepSecretCode() {
		x.StepName = "password"
	}
}

// IsStepID returns true if StepName is empty.
func (x *SignupDTO) IsStepID() bool {
	return x.StepName == ""
}

// IsStepSecretCode returns true if StepName is "SecretCode".
func (x *SignupDTO) IsStepSecretCode() bool {
	return x.StepName == "secret_code"
}

// IsStepPassword returns true if StepName is "Password".
func (x *SignupDTO) IsStepPassword() bool {
	return x.StepName == "password"
}

func (x *AccountSignupController) createDTO() error {

	x.dto = &SignupDTO{}

	dto := x.dto
	c := x.webCtxt

	// fix binding problem (POST,GET,query)
	dto.ReturnURL = c.QueryParam("return_url")

	if x.IsGET {

	} else if x.IsPOST {

		if err := c.Bind(&dto.SignupFormDTO); err != nil {
			return err
		}

		{

			x.validateFields() // basic validation after UnMarshal
		}
	}

	return nil
}

func (x *AccountSignupController) handleDTO() error {

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

		var user *service.UserAccount
		var err error

		gotoNextStep := false

		userCanSignup := false /*Sign up*/

		Email := dto.Email

		isInputValid := dto.IsModelValid()

		if isInputValid {

			user, err = accountService.FindByNormalizedEmail(Email)

			if err != nil {
				return err // error e.g. db connection problem
			}

			if user == nil {
				userCanSignup = true // no user with this Email
			} else {
				dto.AddError("", userLang.Lang("The user already exists. Please use the sign-in page." /*Lang*/))
			}

		}

		if userCanSignup { /*Sign up*/

			if botLimit.LimitSignupActivity(Email) {
				x.status = http.StatusTooManyRequests
				return nil
			}

			switch {
			case dto.IsStepID():
				{
					gotoNextStep = true
					sendSecretMsg := true

					if sendSecretMsg {

						if botLimit.LimitSignupMessage(Email) {
							x.status = http.StatusTooManyRequests
							return nil
						}

						secretCode, err := accountService.GeneratePasscodeConfirmEmail(Email, user)

						if err != nil {
							return err // error e.g. vault connect problem
						}

						x.appService.Messenger().SendSecretCodeToEmail(secretCode, Email, userLang.LangCode())
					}
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
			case dto.IsStepPassword():
				{

					createUserAccount, err := accountService.ValidateTokenConfirmEmail(dto.SecretString, Email, user)

					if err != nil {
						return err // error e.g. vault connect problem
					}

					if createUserAccount {

						user, err := service.NewUserAccount()
						if err != nil {
							return err // error e.g. db connection problem
						}

						user.SetEmail(Email)

						if u, e := accountService.FindByNormalizedEmail(user.Email); u != nil || e != nil {

							if e != nil {
								return e // error e.g. db connection problem
							}

							if u != nil {
								// user exists
								// some sort of misuse of arguments or a collision

								return fmt.Errorf("user with this email exists")
							}

						}

						err = user.SetPassword(dto.Password)
						if err != nil {
							return err
						}

						err = accountService.CreateUserAccount(user)

						if err != nil {
							return err // error e.g. db connection problem
						}

						dto.IsSuccess = true
						dto.IsRenderModeMessage = true
						dto.StatusMessage = userLang.Lang("Thank you for signing up!" /*Lang*/)

						if dto.ReturnURL != "" {
							dto.NoRender = true
							err = c.Redirect(http.StatusFound /*302*/, dto.ReturnURL)
							if err != nil {
								return err
							}
						}

						// TODO sign in force
						// TODO return goto ReturnURL
						// return LocalRedirect(ReturnURL ?? "~/")
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
func (x *AccountSignupController) responseDTOAsAPI() (err error) {

	dto := x.dto

	c := x.webCtxt

	if x.status == 0 {
		x.status = http.StatusOK
	}
	return c.JSON(x.status, dto)

}

func (x *AccountSignupController) responseDTO() (err error) {

	return x.responseDTOAsAPI()

}
