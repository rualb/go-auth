package account

import (
	"go-auth/internal/config"
	"go-auth/internal/config/consts"
	"go-auth/internal/controller"

	"go-auth/internal/i18n"
	"go-auth/internal/mvc"
	"go-auth/internal/service"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

type DeleteDataController struct {
	appService service.AppService

	appConfig *config.AppConfig
	userLang  i18n.UserLang

	IsGET  bool
	IsPOST bool

	webCtxt echo.Context // webCtxt

	userAccount *service.UserAccount

	dto    *DeleteDataDTO
	status int
}

func (x *DeleteDataController) Handler() error {

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
func NewDeleteDataController(appService service.AppService, c echo.Context) *DeleteDataController {

	appConfig := appService.Config()

	return &DeleteDataController{

		appService: appService,

		appConfig:   appConfig,
		userLang:    controller.UserLang(c, appService),
		IsGET:       controller.IsGET(c),
		IsPOST:      controller.IsPOST(c),
		userAccount: controller.GetAccountWithService(c, appService),
		webCtxt:     c,
	}
}

type DeleteDataFormDTO struct {
	Keyword  string `form:"keyword" json:"keyword,omitempty"`
	Password string `form:"password" json:"password,omitempty"`
}
type DeleteDataDTO struct {
	mvc.ModelBaseDTO
	DeleteDataFormDTO

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

func (x *DeleteDataController) validateFields() {
	lang := x.userLang
	dto := x.dto

	{

		dto.Password = strings.TrimSpace(dto.Password)
		dto.Keyword = strings.TrimSpace(dto.Keyword)

	}

	{
		v := dto.NewModelValidatorStr(lang, "keyword", "Keyword" /*Lang*/, dto.Keyword, consts.DefaultTextLength)
		v.Required()
		v.Keyword("delete") // keyword
	}

	{
		v := dto.NewModelValidatorStr(lang, "password", "Password" /*Lang*/, dto.Password, consts.PasswordMaxLength)
		v.Required()
		// v.Password(consts.PasswordMinLength)

	}

}

func (x *DeleteDataController) createDTO() error {

	x.dto = &DeleteDataDTO{}
	//

	dto := x.dto
	c := x.webCtxt

	// fix binding problem (POST,GET,query)
	dto.NextURL = c.QueryParam("next")

	if x.IsGET {

	} else if x.IsPOST {

		if err := c.Bind(&dto.DeleteDataFormDTO); err != nil {
			return err
		}

		{

			x.validateFields() // basic validation after UnMarshal
		}
	}

	return nil
}

func (x *DeleteDataController) handleDTO() error {

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

		isInputValid := dto.IsModelValid()

		if isInputValid {

			var err error

			user := x.userAccount
			if user == nil {
				dto.AddError("", userLang.Lang("No user found." /*Lang*/))
				return nil
			}

			if botLimit.LimitAccountAccess(user.ID) {
				x.status = http.StatusTooManyRequests
				return nil
			}

			if user.CompareHashAndPassword(dto.Password) {
				// set new password
				err = user.SetPassword("") // set empty password
				// mark for deletion

				if err != nil {
					return err
				}

				err = accountService.UpdateUserAccount(user)

				if err != nil {
					return err // error e.g. db connection problem
				}

				{
					// SignOut
					signInService.SignOut()
				}

				// redirect in api page
				// {
				// 	err = c.Redirect(http.StatusFound /*302*/, "/") // go home page
				// 	if err != nil {
				// 		return err
				// 	}
				// }
				dto.Password = ""
				dto.Status = "success"
				// dto.IsSuccess = true
				// dto.IsRenderModeMessage = true
				dto.Message /*StatusMessage*/ = userLang.Lang("Your account has been deleted." /*Lang*/)

			} else {
				dto.AddError("", userLang.Lang("Invalid password." /*Lang*/))
				return nil
			}

		}

	}

	return nil
}

func (x *DeleteDataController) responseDTO() (err error) {

	c := x.webCtxt
	dto := x.dto

	if x.status == 0 {
		x.status = http.StatusOK
	}
	return c.JSON(x.status, dto)
}
