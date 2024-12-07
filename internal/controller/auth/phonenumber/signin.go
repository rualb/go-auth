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

type AccountSigninController struct {
	appService service.AppService

	appConfig *config.AppConfig
	userLang  i18n.UserLang

	IsGET  bool
	IsPOST bool

	webCtxt echo.Context // webCtxt

	dto    *SigninDTO
	status int
}

func (x *AccountSigninController) Handler() error {

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
func NewAccountSigninController(appService service.AppService, c echo.Context) *AccountSigninController {

	appConfig := appService.Config()

	return &AccountSigninController{

		appService: appService,

		appConfig: appConfig,
		userLang:  controller.UserLang(c, appService),
		IsGET:     controller.IsGET(c),
		IsPOST:    controller.IsPOST(c),
		webCtxt:   c,
	}
}

type SigninFormDTO struct {
	PhoneNumber string `form:"phone_number" json:"phone_number,omitempty"`
	Password    string `form:"password" json:"password,omitempty"`
}
type SigninDTO struct {
	mvc.ModelBaseDTO
	SigninFormDTO

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

func (x *AccountSigninController) validateFields() {
	lang := x.userLang
	dto := x.dto

	{

		dto.PhoneNumber = utilstring.NormalizePhoneNumber(dto.PhoneNumber)
		dto.Password = strings.TrimSpace(dto.Password)

	}

	{
		v := dto.NewModelValidatorStr(lang, "phone_number", "Phone number" /*Lang*/, dto.PhoneNumber, consts.DefaultTextLength)
		v.Required()
		v.PhoneNumber()
		v.LengthRange(consts.PhoneNumberMinLength, consts.PhoneNumberMaxLength)

	}

	{
		v := dto.NewModelValidatorStr(lang, "password", "Password" /*Lang*/, dto.Password, consts.PasswordMaxLength)
		v.Required()
		// v.Password(consts.PasswordMinLength) // for signin password check not required

	}

}

func (x *AccountSigninController) createDTO() error {

	x.dto = &SigninDTO{}
	//

	dto := x.dto
	c := x.webCtxt

	// fix binding problem (POST,GET,query)
	dto.ReturnURL = c.QueryParam("return_url")

	if x.IsGET {

	} else if x.IsPOST {

		if err := c.Bind(&dto.SigninFormDTO); err != nil {
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

func (x *AccountSigninController) handleDTO() error {

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

		userExists := false
		userCanSignIn := false /*Sign in*/

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

		if userExists && userCanSignIn { /*Sign in*/

			if botLimit.LimitAccountAccess(user.ID) {
				x.status = http.StatusTooManyRequests
				return nil
			}

			success, err := signInService.PasswordSignIn(user, dto.Password)

			if err != nil {
				return err // error e.g. crypto
			}

			if success {

				if dto.ReturnURL != "" {
					err = c.Redirect(http.StatusFound /*302*/, dto.ReturnURL)
					if err != nil {
						return err //
					}
				}

				dto.IsSuccess = true
				dto.IsRenderModeMessage = true
				dto.StatusMessage = userLang.Lang("You have successfully signed in." /*Lang*/)

			} else {
				dto.AddError("", userLang.Lang("Invalid sign in attempt." /*Lang*/))

			}

		}

	}

	return nil
}

func (x *AccountSigninController) responseDTOAsAPI() (err error) {

	c := x.webCtxt
	dto := x.dto

	if x.status == 0 {
		x.status = http.StatusOK
	}
	return c.JSON(x.status, dto)

}

func (x *AccountSigninController) responseDTO() (err error) {

	return x.responseDTOAsAPI()

}
