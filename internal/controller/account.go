package controller

// import (
// 	"go-auth/internal/container"
// )

// // AccountController is a controller for managing user account.
// type AccountController interface {
// }

// type accountController struct {
// 	container container.AppContainer
// 	// accountService service.AccountService
// 	// vaultServiceAuth   service.VaultService
// 	// vaultServiceSignup service.VaultService
// 	// anonymousUser  *dto.UserIdDTO

// }

// // NewAccountController is constructor.
// func NewAccountController(container container.AppContainer) AccountController {

// 	return &accountController{
// 		container: container,
// 		// accountService: service.NewAccountService(container),
// 		// vaultServiceAuth:   service.NewVaultService(container),
// 		// vaultServiceSignup: service.NewVaultService(container, ),
// 		// anonymousUser:  model.NewUserIdDTO(""),

// 	}
// }

// type UserLoginDTO struct {
// 	Password string `form:"Password"`
// 	Username string `form:"Username"`
// }

// func (controller *accountController) AccountInfo(c echo.Context) error {
// 	// SecurityEnabled := true
// 	// if !SecurityEnabled {
// 	// 	return c.JSON(http.StatusOK, controller.anonymousUser)
// 	// }

// 	acc, err := controller.Account(c)

// 	if err != nil {
// 		return err
// 	}
// 	if acc == nil {
// 		return c.NoContent(http.StatusNotFound)
// 	}
// 	return c.JSON(http.StatusOK, acc)
// }

// func (controller *accountController) SigninGET(c echo.Context) error {
// 	dtoModel := &dto.UserLoginDTO{Username: "Undefined"}

// 	data := mvc.NewModelWrap(c, dtoModel, true, "Sign in")
// 	err := c.Render(http.StatusOK, "signin.html", data)
// 	return err
// }

// func (controller *accountController) SigninPOST(c echo.Context) error {

// 	dtoModel := &dto.UserLoginDTO{}

// 	if err := c.Bind(dtoModel); err != nil {
// 		return err
// 	}

// 	model := mvc.NewModelWrap(c, dtoModel, true, "Sign in")
// 	err := c.Render(http.StatusOK, "signin.html", model)
// 	return err
// }

// func (controller *accountController) Signin(c echo.Context) error {
// 	obj := dto.NewUserLoginDTO()
// 	if err := c.Bind(obj); err != nil {
// 		return c.JSON(http.StatusBadRequest, obj)
// 	}

// 	a, err := controller.accountService.FindByUsernameAndPassword(obj.Username, obj.Password)
// 	if err != nil {
// 		return err
// 	}

// 	if a != nil {

// 		claims := token.TokenClaims(controller.container)
// 		// claims.AddScope(token.ScopeAuth)
// 		claims.UserID = a.ID
// 		err := token.CreateAuthTokenWithClaims(c, claims, controller.vaultServiceAuth)
// 		if err != nil {
// 			return err
// 		}

// 		c.NoContent(http.StatusOK)
// 	}

// 	return c.NoContent(http.StatusUnauthorized)
// }

// func (controller *accountController) SignOut(c echo.Context) error {

// 	token.DeleteAuthToken(c)

// 	return c.NoContent(http.StatusOK)
// }

// func (controller *accountController) Account(c echo.Context) (*model.UserAccount, error) {

// 	accountRaw := c.Get("account")

// 	if accountRaw == nil {
// 		id := token.UserID(c)
// 		account, err := controller.accountService.FindByID(id)
// 		if err != nil {
// 			return nil, err
// 		}

// 		accountRaw = account
// 		c.Set("account", account) // cache
// 	}

// 	if accountRaw != nil {
// 		return accountRaw.(*model.UserAccount), nil
// 	}

// 	return nil, nil
// }
