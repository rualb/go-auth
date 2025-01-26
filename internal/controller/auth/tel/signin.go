package tel

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
	Tel      string `form:"tel" json:"tel,omitempty"`
	Password string `form:"password" json:"password,omitempty"`
}
type SigninDTO struct {
	mvc.ModelBaseDTO
	SigninFormDTO

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

func (x *AccountSigninController) validateFields() {
	lang := x.userLang
	dto := x.dto

	{

		dto.Tel = utilstring.NormalizeTel(dto.Tel)
		dto.Password = strings.TrimSpace(dto.Password)

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
		// v.Password(consts.PasswordMinLength) // for signin password check not required

	}

}

func (x *AccountSigninController) createDTO() error {

	x.dto = &SigninDTO{}
	//

	dto := x.dto
	c := x.webCtxt

	// fix binding problem (POST,GET,query)
	dto.NextURL = c.QueryParam("next")

	if x.IsGET {

	} else if x.IsPOST {

		if err := c.Bind(&dto.SigninFormDTO); err != nil {
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

			user, err = accountService.FindByTel(dto.Tel)

			if err != nil {
				return err // error e.g. db connection problem
			}

			if user == nil {
				dto.AddError("", userLang.Lang("No user found." /*Lang*/))
			} else {
				userExists = true
				userCanSignIn = signInService.CanSignIn(user) // no user with this Tel

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

				if dto.NextURL != "" {
					err = c.Redirect(http.StatusFound /*302*/, dto.NextURL)
					if err != nil {
						return err //
					}
				}

				dto.Password = ""
				dto.Status = "success"
				// dto.IsSuccess = true
				// dto.IsRenderModeMessage = true
				dto.Message /*StatusMessage*/ = userLang.Lang("You have successfully signed in." /*Lang*/)

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
