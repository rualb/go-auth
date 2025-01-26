package tel

import (
	"fmt"
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
	Tel string `form:"tel" json:"tel,omitempty"`

	Password string `form:"password" json:"password,omitempty"`

	Passcode string `form:"passcode" json:"passcode,omitempty"`
	Step     string `form:"step" json:"step,omitempty"`
	Token    string `form:"token" json:"token,omitempty"`
}

type SignupDTO struct {
	mvc.ModelBaseDTO
	SignupFormDTO
	//
	// StateString  string `form:"state_string"` //=><input value="{{ .Model.StateString }}" name="StateString" type="hidden" />
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

func (x *AccountSignupController) validateFields() {

	lang := x.userLang
	dto := x.dto

	{

		dto.Tel = utilstring.NormalizeTel(dto.Tel)
		dto.Password = strings.TrimSpace(dto.Password)

		dto.Passcode = strings.TrimSpace(dto.Passcode)

	}

	{
		v := dto.NewModelValidatorStr(lang, "tel", "Phone number" /*Lang*/, dto.Tel, consts.DefaultTextLength)
		v.Required()
		v.Tel()
		v.LengthRange(consts.TelMinLength, consts.TelMaxLength)

	}

	{
		v := dto.NewModelValidatorStr(lang, "password", "Password" /*Lang*/, dto.Password, consts.PasswordMaxLength)
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
			dto.RemoveError("password")
		}

		if dto.IsStepPasscode() {
			// dto.RemoveError("passcode")
			dto.RemoveError("password")
		}

		if dto.IsStepPassword() {
			dto.RemoveError("passcode")
		}
	}
}

func (x *SignupDTO) StepNext() {
	if x.IsStepID() {
		x.Step = "passcode"
	} else if x.IsStepPasscode() {
		x.Step = "password"
	}
}

// IsStepID returns true if Step is empty.
func (x *SignupDTO) IsStepID() bool {
	return x.Step == ""
}

// IsStepPasscode returns true if Step is "passcode".
func (x *SignupDTO) IsStepPasscode() bool {
	return x.Step == "passcode"
}

// IsStepPassword returns true if Step is "password".
func (x *SignupDTO) IsStepPassword() bool {
	return x.Step == "password"
}

func (x *AccountSignupController) createDTO() error {

	x.dto = &SignupDTO{}

	dto := x.dto
	c := x.webCtxt

	// fix binding problem (POST,GET,query)
	dto.NextURL = c.QueryParam("next")

	if x.IsGET {

	} else if x.IsPOST {

		if err := c.Bind(&dto.SignupFormDTO); err != nil {
			return err
		}

		{

			x.validateFields() // basic validation after UnMarshal
		}
	}

	if dto.Tel == "" {
		dto.Tel = x.appConfig.Identity.TelPrefix
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

		isInputValid := dto.IsModelValid()

		if isInputValid {

			user, err = accountService.FindByTel(dto.Tel)

			if err != nil {
				return err // error e.g. db connection problem
			}

			if user == nil {
				userCanSignup = true // no user with this Tel
			} else {
				dto.AddError("", userLang.Lang("The user already exists. Please use the sign-in page." /*Lang*/))
			}

		}

		if userCanSignup { /*Sign up*/

			if botLimit.LimitSignupActivity(dto.Tel) {
				x.status = http.StatusTooManyRequests
				return nil
			}

			switch {
			case dto.IsStepID():
				{
					gotoNextStep = true
					sendSecretMsg := true

					if sendSecretMsg {

						if botLimit.LimitSignupMessage(dto.Tel) {
							x.status = http.StatusTooManyRequests
							return nil
						}

						passcode, err := accountService.GeneratePasscodeConfirmTel(dto.Tel, user)

						if err != nil {
							return err // error e.g. vault connect problem
						}

						x.appService.Messenger().SendPasscodeToTel(passcode, dto.Tel, userLang.LangCode())
					}

				}
			case dto.IsStepPasscode():
				{

					isPasscodeValid, err := accountService.ValidatePasscodeConfirmTel(dto.Passcode, dto.Tel, user)

					if err != nil {
						return err // error e.g. vault connect problem
					}
					if isPasscodeValid {
						{
							token, err := accountService.GenerateTokenConfirmTel(dto.Tel, user)
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
			case dto.IsStepPassword():
				{

					createUserAccount, err := accountService.ValidateTokenConfirmTel(dto.Token, dto.Tel, user)

					if err != nil {
						return err // error e.g. vault connect problem
					}

					if createUserAccount {

						user, err := service.NewUserAccount()
						if err != nil {
							return err // error e.g. db connection problem
						}

						user.SetTel(dto.Tel)

						if u, e := accountService.FindByTel(user.Tel); u != nil || e != nil {

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

						dto.Password = ""
						dto.Token = ""
						dto.Passcode = ""
						dto.Status = "success"
						// dto.IsSuccess = true
						// dto.IsRenderModeMessage = true
						dto.Message /*StatusMessage*/ = userLang.Lang("Thank you for signing up!" /*Lang*/)

						if dto.NextURL != "" {
							dto.NoRender = true
							err = c.Redirect(http.StatusFound /*302*/, dto.NextURL)
							if err != nil {
								return err
							}
						}

						// TODO sign in force
						// TODO return goto NextURL
						/// return LocalRedirect(NextURL ?? "~/")
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

	// TODO maybe set password-passcode "" an empty string on return

	return x.responseDTOAsAPI()

}
