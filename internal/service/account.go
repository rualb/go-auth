package service

import (
	"fmt"
	xtoken "go-auth/internal/token"
	"go-auth/internal/tool/toolcrypto"
	"go-auth/internal/tool/toolstring"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	IssuerConfirmPhoneNumber = "confirm_phone"
	IssuerConfirmEmail       = "confirm_email"
)
const (
	TokenLifetimeSignupWithPhoneNumber = time.Minute * 30 // 30 minutes

	TokenLifetimeSignupWithEmail = time.Minute * 30 // 30 minutes
)
const (
	SecurityStampLenDefault = 16
)

// UserAccount Username,Email,NormalizedEmail are uniqueIndex with condition "not empty"
type UserAccount struct {
	ID              string `gorm:"primaryKey"`
	Username        string `gorm:"uniqueIndex:,where:username != ''"`
	PhoneNumber     string
	Email           string // use this on emailing and show
	NormalizedEmail string `gorm:"uniqueIndex:,where:normalized_email != ''"` // use this on search
	// SecurityStamp   string // Key := Base32(Random(32))  HMACSHA1(Key)  Key == VTOQQ2PQKD7A2KTSXU7OFLKUNI7QEZRJ
	PasswordHash string
	CreatedAt    time.Time
	Role         string
}

func (x *UserAccount) SetUsername(value string) {
	valueNorm := toolstring.NormalizeText(value)
	x.Username = valueNorm
}

func (x *UserAccount) SetPhoneNumber(value string) {
	valueNorm := toolstring.NormalizePhoneNumber(value)
	x.PhoneNumber = valueNorm

	// x.Username = valueNorm
}

func (x *UserAccount) SetEmail(value string) {
	valueNorm := toolstring.NormalizeEmail(value)

	x.Email = value
	x.NormalizedEmail = valueNorm

	// x.Username = valueNorm
}

func (x *UserAccount) SetPassword(pw string) error {

	hash, err := toolcrypto.HashPassword(pw) // bcrypt inside

	if err != nil {
		return err
	}

	x.PasswordHash = hash

	// x.RefreshSecurityStamp()

	return nil
}
func (x *UserAccount) CompareHashAndPassword(str string) bool {

	return toolcrypto.CompareHashAndPassword(x.PasswordHash, str)

}

// func (x *UserAccount) RefreshSecurityStamp() error {
// 	stamp, e := toolcrypto.RandomCryptoBase32(SecurityStampLenDefault)

// 	if e != nil {
// 		return e
// 	}

// 	x.SecurityStamp = stamp

// 	return nil
// }

func NewUserAccount() (*UserAccount, error) {

	now := time.Now().UTC() // now

	id := uuid.New().String()
	res := &UserAccount{
		CreatedAt: now,
		ID:        id,
	}

	// err := res.RefreshSecurityStamp()
	// if err != nil {
	// 	return nil, err
	// }

	return res, nil
}

func makeScope(scopeName string, scopeUser *UserAccount, scopeInfo string) string {

	arr := make([]string, 0, 3)

	if scopeInfo != "" {
		arr = append(arr, scopeName)
	}

	if scopeUser != nil && scopeUser.ID != "" {
		arr = append(arr, scopeUser.ID)
	}

	if scopeInfo != "" {
		arr = append(arr, scopeInfo)
	}

	return strings.Join(arr, "|")
}

// AccountService is a service for managing user account.
type AccountService interface {
	CanSignInWithPassword(userAccount *UserAccount, password string) (passwordValid bool)

	FindByID(id string) (*UserAccount, error)
	FindByUsername(name string) (*UserAccount, error)
	FindByPhoneNumber(name string) (*UserAccount, error)
	FindByNormalizedEmail(email string) (*UserAccount, error)
	CreateUserAccount(userAccount *UserAccount) (err error)
	UpdateUserAccount(userAccount *UserAccount) (err error)

	GeneratePasscodeConfirmPhoneNumber(phoneNumber string, userAccount *UserAccount) (string, error)
	ValidatePasscodeConfirmPhoneNumber(passcode string, phoneNumber string, userAccount *UserAccount) (bool, error)

	GeneratePasscodeConfirmEmail(email string, userAccount *UserAccount) (string, error)
	ValidatePasscodeConfirmEmail(passcode string, email string, userAccount *UserAccount) (bool, error)

	GenerateTokenConfirmPhoneNumber(phoneNumber string, userAccount *UserAccount) (secretToken string, err error)
	ValidateTokenConfirmPhoneNumber(secretToken string, phoneNumber string, userAccount *UserAccount) (ok bool, err error)

	GenerateTokenConfirmEmail(email string, userAccount *UserAccount) (secretToken string, err error)
	ValidateTokenConfirmEmail(secretToken string, email string, userAccount *UserAccount) (ok bool, err error)
}

type defaultAccountService struct {
	appService AppService
}

// GenerateTokenConfirmPhoneNumber create token for trust phone number data
func (x *defaultAccountService) GenerateTokenConfirmPhoneNumber(phoneNumber string, userAccount *UserAccount) (secretToken string, err error) {

	if phoneNumber == "" {
		return "", fmt.Errorf("error: phone number is empty")
	}
	if !toolstring.IsPhoneNumberFull(phoneNumber) {
		return "", fmt.Errorf("error: not a valid phone number: %v", phoneNumber)
	}

	vaultKeyScopeHash := x.appService.Vault().KeyScopeHash()
	scope := makeScope(IssuerConfirmPhoneNumber, userAccount, phoneNumber)

	claims := &xtoken.TokenClaimsDTO{}
	claims.SetIssuer(IssuerConfirmPhoneNumber)
	claims.AddScope(scope)                                 // step 2
	claims.SetLifetime(TokenLifetimeSignupWithPhoneNumber) // step 3
	token, err := xtoken.CreateToken(claims, vaultKeyScopeHash /*Signup vault*/)

	if err != nil {
		return "", err
	}

	return token, nil
}

// VerifyConfirmPhoneNumberOtpToken validate token for trust phone number data
func (x *defaultAccountService) ValidateTokenConfirmPhoneNumber(secretToken string, phoneNumber string, userAccount *UserAccount) (ok bool, err error) {

	if phoneNumber == "" {
		return false, fmt.Errorf("error: phone number is empty")
	}
	if !toolstring.IsPhoneNumberFull(phoneNumber) {
		return false, fmt.Errorf("error: not a valid phone number: %v", phoneNumber)
	}
	// // // // //

	vaultKeyScopeHash := x.appService.Vault().KeyScopeHash()

	claims, err := xtoken.ParseToken(secretToken, vaultKeyScopeHash /*Signup vault*/)

	if err != nil {
		return false, err
	}

	scope := makeScope(IssuerConfirmPhoneNumber, userAccount, phoneNumber)

	// valid and has defined scope
	if claims != nil &&
		claims.IsValid() &&
		claims.HasScope(scope) &&
		claims.IsIssuedBy(IssuerConfirmPhoneNumber) {
		return true, nil // valid
	}

	return ok, nil
}

// VerifyConfirmPhoneNumberOtpToken implements AccountService.
func (x *defaultAccountService) ValidatePasscodeConfirmPhoneNumber(passcode string, phoneNumber string, userAccount *UserAccount) (bool, error) {

	if phoneNumber == "" {
		return false, fmt.Errorf("error: phone number is empty")
	}
	if !toolstring.IsPhoneNumberFull(phoneNumber) {
		return false, fmt.Errorf("error: not a valid phone number: %v", phoneNumber)
	}
	vaultKeyScopeOtp := x.appService.Vault().KeyScopeOtp()
	_, secret, err := vaultKeyScopeOtp.CurrentKey()

	if err != nil {
		return false, err
	}

	scope := makeScope(IssuerConfirmPhoneNumber, userAccount, phoneNumber)

	config := xtoken.NewConfigTotp(scope, secret)

	ok, err := xtoken.ValidatePasscode(passcode, config)

	if err != nil {
		return false, err
	}

	return ok, nil
}

// GeneratePasscodeConfirmPhoneNumber implements AccountService.
func (x *defaultAccountService) GeneratePasscodeConfirmPhoneNumber(phoneNumber string, userAccount *UserAccount) (string, error) {

	if phoneNumber == "" {
		return "", fmt.Errorf("error: phone number is empty")
	}
	if !toolstring.IsPhoneNumberFull(phoneNumber) {
		return "", fmt.Errorf("error: not a valid phone number: %v", phoneNumber)
	}

	vaultKeyScopeOtp := x.appService.Vault().KeyScopeOtp()
	_, secret, err := vaultKeyScopeOtp.CurrentKey()

	if err != nil {
		return "", err
	}

	scope := makeScope(IssuerConfirmPhoneNumber, userAccount, phoneNumber)

	config := xtoken.NewConfigTotp(scope, secret)

	passcode, err := xtoken.GeneratePasscode(config)

	if err != nil {
		return "", err
	}

	return passcode, nil
}

// GenerateTokenConfirmEmail create token for email
func (x *defaultAccountService) GenerateTokenConfirmEmail(email string, userAccount *UserAccount) (secretToken string, err error) {

	if email == "" {
		return "", fmt.Errorf("error: email is empty")
	}
	if !toolstring.IsEmail(email) {
		return "", fmt.Errorf("error: not a valid email: %v", email)
	}

	scope := makeScope(IssuerConfirmEmail, userAccount, email)

	vaultKeyScopeHash := x.appService.Vault().KeyScopeHash()

	claims := &xtoken.TokenClaimsDTO{}
	// 3 steps - phone,scope and lifetime
	claims.SetIssuer(IssuerConfirmEmail)
	claims.AddScope(scope)                           // step 2
	claims.SetLifetime(TokenLifetimeSignupWithEmail) // step 3
	token, err := xtoken.CreateToken(claims, vaultKeyScopeHash /*Signup vault*/)

	if err != nil {
		return "", err
	}

	return token, nil
}

// VerifyConfirmEmailOtpToken validate token for email
func (x *defaultAccountService) ValidateTokenConfirmEmail(secretToken string, email string, userAccount *UserAccount) (ok bool, err error) {
	if email == "" {
		return false, fmt.Errorf("error: email is empty")
	}
	if !toolstring.IsEmail(email) {
		return false, fmt.Errorf("error: not a valid email: %v", email)
	}

	vaultKeyScopeHash := x.appService.Vault().KeyScopeHash()

	claims, err := xtoken.ParseToken(secretToken, vaultKeyScopeHash /*Signup vault*/)

	if err != nil {
		return false, err
	}

	scope := makeScope(IssuerConfirmEmail, userAccount, email)

	// valid and has defined scope
	if claims != nil &&
		claims.IsValid() &&
		claims.HasScope(scope) &&
		claims.IsIssuedBy(IssuerConfirmEmail) {
		return true, nil // valid
	}

	return ok, nil
}

// VerifyConfirmEmailOtpToken implements AccountService.
func (x *defaultAccountService) ValidatePasscodeConfirmEmail(passcode string, email string, userAccount *UserAccount) (bool, error) {

	if email == "" {
		return false, fmt.Errorf("error: email is empty")
	}
	if !toolstring.IsEmail(email) {
		return false, fmt.Errorf("error: not a valid email: %v", email)
	}

	vaultKeyScopeOtp := x.appService.Vault().KeyScopeOtp()
	_, secret, err := vaultKeyScopeOtp.CurrentKey()

	if err != nil {
		return false, err
	}

	scope := makeScope(IssuerConfirmEmail, userAccount, email)

	config := xtoken.NewConfigTotp(scope, secret)

	ok, err := xtoken.ValidatePasscode(passcode, config)

	if err != nil {
		return false, err
	}

	return ok, nil
}

// GeneratePasscodeConfirmEmail implements AccountService.
func (x *defaultAccountService) GeneratePasscodeConfirmEmail(email string, userAccount *UserAccount) (string, error) {

	if email == "" {
		return "", fmt.Errorf("error: email is empty")
	}
	if !toolstring.IsEmail(email) {
		return "", fmt.Errorf("error: not a valid email: %v", email)
	}

	vaultKeyScopeOtp := x.appService.Vault().KeyScopeOtp()
	_, secret, err := vaultKeyScopeOtp.CurrentKey()

	if err != nil {
		return "", err
	}

	scope := makeScope(IssuerConfirmEmail, userAccount, email)

	config := xtoken.NewConfigTotp(scope, secret)

	passcode, err := xtoken.GeneratePasscode(config)

	if err != nil {
		return "", err
	}

	return passcode, nil
}

// NewAccountService is constructor.
func newAccountService(appService AppService) AccountService {

	return &defaultAccountService{
		appService: appService,
	}
}

// find by using username and plain text password.
func (x defaultAccountService) CanSignInWithPassword(userAccount *UserAccount, password string) (passwordValid bool) {

	passwordValid = false

	if x.canSignIn(userAccount) && userAccount.CompareHashAndPassword(password) {
		passwordValid = true
	}

	return passwordValid
}

func (x defaultAccountService) canSignIn(userAccount *UserAccount) bool {

	return userAccount != nil && userAccount.PasswordHash != ""
}

func (x defaultAccountService) CreateUserAccount(userAccount *UserAccount) (err error) {
	result := x.appService.Repository().Create(userAccount)

	if result.Error != nil {
		return result.Error
	}

	return nil
}
func (x defaultAccountService) UpdateUserAccount(userAccount *UserAccount) (err error) {
	// result := x.appService.Repository().Updates(userAccount)
	result := x.appService.Repository().Save(userAccount)

	if result.Error != nil {
		return result.Error
	}

	return nil
}

// // insert new
// func (x accountService) Create(m *UserAccount) error {

// 	result := x.appService.Repository().Create(m)

// 	return result.Error
// }

// // update if exists
// func (x accountService) Update(m *UserAccount) error {

// 	result := x.appService.Repository().Updates(m)

// 	return result.Error
// }

func (x defaultAccountService) FindByID(id string) (*UserAccount, error) {

	if id == "" {
		return nil, nil // fmt.Errorf("id cannot be empty")
	}

	user := new(UserAccount)

	result := x.appService.Repository().Find(user, "id = ?", id)

	if result.Error != nil || result.RowsAffected == 0 {
		return nil, result.Error
	}

	return user, nil
}
func (x defaultAccountService) FindByUsername(username string) (*UserAccount, error) {

	if username == "" {
		return nil, nil
	}

	username = toolstring.NormalizeText(username)

	user := new(UserAccount)

	result := x.appService.Repository().Find(user, "username = ?", username)

	if result.Error != nil || result.RowsAffected == 0 {
		return nil, result.Error
	}

	return user, nil
}
func (x defaultAccountService) FindByPhoneNumber(phoneNumber string) (*UserAccount, error) {
	if phoneNumber == "" {
		return nil, nil
	}

	if !toolstring.IsPhoneNumberFull(phoneNumber) {
		return nil, nil
	}

	user := new(UserAccount)

	result := x.appService.Repository().Find(user, "phone_number = ?", phoneNumber)

	if result.Error != nil || result.RowsAffected == 0 {
		return nil, result.Error
	}

	return user, nil
}
func (x defaultAccountService) FindByNormalizedEmail(email string) (*UserAccount, error) {

	if email == "" {
		return nil, nil
	}

	if !toolstring.IsEmail(email) {
		return nil, nil
	}

	normalizedEmail := toolstring.NormalizeEmail(email)

	user := new(UserAccount)

	result := x.appService.Repository().Find(user, "normalized_email = ?", normalizedEmail)

	if result.Error != nil || result.RowsAffected == 0 {
		return nil, result.Error
	}

	return user, nil
}
