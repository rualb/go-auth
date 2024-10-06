package service

import (
	"fmt"
	"go-auth/internal/config"
	xtoken "go-auth/internal/token"
	"time"
)

type SignInService interface {
	CanSignIn(userAccount *UserAccount) (accountValid bool)
	PasswordSignIn(userAccount *UserAccount, password string) (success bool, err error)
	SignIn(userAccount *UserAccount) error
	SignOut()
	IsSignedIn() (accountSignedIn bool)

	// if (SignInManager.IsSignedIn(User))
}

// NewSignInService is constructor.
func NewSignInService(appService AppService, tokenPersist xtoken.TokenPersist) SignInService {
	return &defaultSigninService{
		tokenPersist: tokenPersist,
		appService:   appService,
	}
}

type defaultSigninService struct {
	tokenPersist xtoken.TokenPersist
	appService   AppService
}

func (x *defaultSigninService) CanSignIn(userAccount *UserAccount) (accountValid bool) {
	return userAccount != nil // && userAccount.PasswordHash != ""
}

func (x *defaultSigninService) PasswordSignIn(userAccount *UserAccount, password string) (success bool, err error) {

	if !x.CanSignIn(userAccount) {
		return false, nil
	}

	if userAccount.CompareHashAndPassword(password) {
		err := x.SignIn(userAccount)

		if err == nil {
			return true, nil

		} else {
			// TODO
			return false, fmt.Errorf("password sing in: %v", err)
		}

	}

	return false, nil
}

func (x *defaultSigninService) SignIn(userAccount *UserAccount) error {

	claims := newTokenClaims(x.appService.Config())
	// claims.AddScope(xtoken.ScopeAuth)
	claims.UserID = userAccount.ID
	err := x.tokenPersist.CreateAuthTokenWithClaims(claims)
	if err != nil {
		return err
	}

	return nil
}

// RotateSignIn rotate token, extend lifetime
func (x *defaultSigninService) RotateSignIn(forceRotate bool) {

	x.tokenPersist.RotateAuthToken(forceRotate)

}

func (x *defaultSigninService) SignOut() {

	x.tokenPersist.DeleteAuthToken()
}

func (x *defaultSigninService) IsSignedIn() (accountSignedIn bool) {

	claims := x.tokenPersist.AuthTokenClaims()

	if claims != nil && claims.IsSignedIn() {

		return true
	}

	return false

}

func newTokenClaims(config *config.AppConfig) *xtoken.TokenClaimsDTO {

	configToken := config.Identity
	claims := &xtoken.TokenClaimsDTO{}

	claims.SetIssuer(configToken.AuthTokenIssuer) //!!!

	//
	claims.SetLifetime(time.Duration(configToken.TokenMaxAge) * time.Second)

	return claims

}
