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

type AccountSigninController struct {
	appService service.AppService

	appConfig *config.AppConfig
	userLang  i18n.UserLang

	IsGET  bool
	IsPOST bool

	webCtxt echo.Context // webCtxt

	isAPIMode bool

	dto *SigninDTO
}

func (x *AccountSigninController) Handler() error {

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
func NewAccountSigninController(appService service.AppService, c echo.Context, isAPIMode bool) *AccountSigninController {

	appConfig := appService.Config()

	return &AccountSigninController{

		appService: appService,
		isAPIMode:  isAPIMode,
		appConfig:  appConfig,
		userLang:   controller.UserLang(c, appService),
		IsGET:      controller.IsGET(c),
		IsPOST:     controller.IsPOST(c),
		webCtxt:    c,
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

		dto.PhoneNumber = toolstring.NormalizePhoneNumber(dto.PhoneNumber)
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
		// v.Password(consts.PasswordMinLength) // for signin password check not required

	}

}

func (x *AccountSigninController) createDto() error {

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

func (x *AccountSigninController) handleDto() error {

	dto := x.dto
	userLang := x.userLang
	c := x.webCtxt

	accountService := x.appService.AccountService()

	signInService := controller.SignInService(c, x.appService)

	if x.IsPOST {

		toolratelimit.RateLimitHuman()

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
				dto.AddModelError("", userLang.Lang("No user found." /*Lang*/))
			} else {
				userExists = true
				userCanSignIn = signInService.CanSignIn(user) // no user with this PhoneNumber

				if !userCanSignIn {
					dto.AddModelError("", userLang.Lang("User account locked out." /*Lang*/))
				}
			}

		}

		if userExists && userCanSignIn { /*Sign in*/
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
				dto.AddModelError("", userLang.Lang("Invalid sign in attempt." /*Lang*/))

			}

		}

	}

	return nil
}

func (x *AccountSigninController) responseDtoAsAPI() (err error) {

	c := x.webCtxt
	dto := x.dto

	controller.CsrfToHeader(c)
	return c.JSON(http.StatusOK, dto)

}

func (x *AccountSigninController) responseDtoAsMvc() (err error) {

	dto := x.dto
	appConfig := x.appConfig
	lang := x.userLang
	c := x.webCtxt

	data, err := mvc.NewModelWrap(c, dto, dto.IsFragment, "Sign in" /*Lang*/, appConfig, lang)
	if err != nil {
		return err
	}
	err = c.Render(http.StatusOK, "signin-phone-number.html", data)
	if err != nil {
		return err
	}

	return nil
}

func (x *AccountSigninController) responseDto() (err error) {

	if x.isAPIMode {
		return x.responseDtoAsAPI()
	} else {
		return x.responseDtoAsMvc()
	}

}
