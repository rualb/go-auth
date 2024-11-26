package phonenumber

import (
	"fmt"
	"go-auth/internal/config"
	"go-auth/internal/config/consts"
	controller "go-auth/internal/controller"
	"go-auth/internal/util/utilratelimit"
	"go-auth/internal/util/utilstring"

	"go-auth/internal/i18n"
	"go-auth/internal/mvc"
	"go-auth/internal/service"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

type AccountSignupController struct {
	appService service.AppService
	appConfig  *config.AppConfig
	userLang   i18n.UserLang

	IsGET  bool
	IsPOST bool

	webCtxt echo.Context // webCtxt

	isAPIMode bool

	dto *SignupDTO
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
func NewAccountSignupController(appService service.AppService, c echo.Context, isAPIMode bool) *AccountSignupController {

	appConfig := appService.Config()

	return &AccountSignupController{

		appService: appService,

		isAPIMode: isAPIMode,
		appConfig: appConfig,
		userLang:  controller.UserLang(c, appService),
		IsGET:     controller.IsGET(c),
		IsPOST:    controller.IsPOST(c),
		webCtxt:   c,
	}
}

type SignupFormDTO struct {
	PhoneNumber string `form:"phone_number" json:"phone_number,omitempty"`

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

		dto.PhoneNumber = utilstring.NormalizePhoneNumber(dto.PhoneNumber)
		dto.Password = strings.TrimSpace(dto.Password)

	}

	{
		v := dto.NewModelValidatorStr(lang, "phone_number", "Phone number" /*Lang*/, dto.PhoneNumber, consts.DefaultTextSize)
		v.Required()
		v.PhoneNumber()
		v.LengthRange(consts.PhoneNumberMinLength, consts.PhoneNumberMaxLength)

	}

	{
		v := dto.NewModelValidatorStr(lang, "password", "Password" /*Lang*/, dto.Password, consts.DefaultTextSize)
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

	if dto.PhoneNumber == "" {
		dto.PhoneNumber = x.appConfig.Identity.PhoneNumberPrefix
	}

	return nil
}

func (x *AccountSignupController) handleDTO() error {

	dto := x.dto

	userLang := x.userLang
	c := x.webCtxt

	accountService := x.appService.AccountService()
	if x.IsPOST {

		utilratelimit.RateLimitHuman()

		var user *service.UserAccount
		var err error

		sendSms := false
		gotoNextStep := false

		userCanSignup := false /*Sign up*/

		isInputValid := dto.IsModelValid()

		if isInputValid {

			user, err = accountService.FindByPhoneNumber(dto.PhoneNumber)

			if err != nil {
				return err // error e.g. db connection problem
			}

			if user == nil {
				userCanSignup = true // no user with this PhoneNumber
			} else {
				dto.AddError("", userLang.Lang("The user already exists. Please use the sign-in page." /*Lang*/))
			}

		}

		if userCanSignup { /*Sign up*/

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
			case dto.IsStepPassword():
				{

					createUserAccount, err := accountService.ValidateTokenConfirmPhoneNumber(dto.SecretString, dto.PhoneNumber, user)

					if err != nil {
						return err // error e.g. vault connect problem
					}

					if createUserAccount {

						user, err := service.NewUserAccount()
						if err != nil {
							return err // error e.g. db connection problem
						}

						user.SetPhoneNumber(dto.PhoneNumber)

						if u, e := accountService.FindByPhoneNumber(user.PhoneNumber); u != nil || e != nil {

							if e != nil {
								return e // error e.g. db connection problem
							}

							if u != nil {
								// user exists
								// some sort of misuse of arguments or a collision

								return fmt.Errorf("user with this phone number exists")
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
						/// return LocalRedirect(ReturnURL ?? "~/")
					}
				}
			}
		}

		if sendSms {
			utilratelimit.RateLimitMessage()

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
func (x *AccountSignupController) responseDTOAsAPI() (err error) {

	dto := x.dto

	c := x.webCtxt

	controller.CsrfToHeader(c)
	return c.JSON(http.StatusOK, dto)

}

func (x *AccountSignupController) responseDTOAsMvc() (err error) {

	dto := x.dto
	appConfig := x.appConfig
	lang := x.userLang
	c := x.webCtxt

	if dto.NoRender {
		return nil
	}

	data, err := mvc.NewModelWrap(c, dto, dto.IsFragment, "Sign up" /*Lang*/, appConfig, lang)
	if err != nil {
		return err
	}
	err = c.Render(http.StatusOK, "signup-phone-number.html", data)
	if err != nil {
		return err
	}

	return nil
}
func (x *AccountSignupController) responseDTO() (err error) {

	// TODO maybe set password-passcode "" an empty string on return

	if x.isAPIMode {
		return x.responseDTOAsAPI()
	} else {
		return x.responseDTOAsMvc()
	}
}
