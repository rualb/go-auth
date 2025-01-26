package service

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewUserAccount(t *testing.T) {

	account, err := NewUserAccount()
	assert.NoError(t, err)
	assert.NotNil(t, account)
	assert.NotEmpty(t, account.ID)
	assert.True(t, account.CreatedAt.Before(time.Now().Add(time.Millisecond)))

	deleteAllAccounts()
}

func TestSetUsername(t *testing.T) {

	account, _ := NewUserAccount()
	account.SetUsername("TestUser")
	assert.Equal(t, "testuser", account.Username)

	deleteAllAccounts()
}

func TestSetTel(t *testing.T) {

	account, _ := NewUserAccount()
	account.SetTel("+123121234567")
	assert.Equal(t, "+123121234567", account.Tel)

	deleteAllAccounts()
}

func TestSetEmail(t *testing.T) {

	account, _ := NewUserAccount()
	account.SetEmail("User@Example.com")
	assert.Equal(t, "User@Example.com", account.Email)
	assert.Equal(t, "user@example.com", account.NormalizedEmail)

	deleteAllAccounts()
}

func TestSetPassword(t *testing.T) {

	account, _ := NewUserAccount()
	err := account.SetPassword("StrongPass1")
	assert.NoError(t, err)
	assert.NotEmpty(t, account.PasswordHash)

	isValid := account.CompareHashAndPassword("StrongPass1")
	assert.True(t, isValid)

	isInvalid := account.CompareHashAndPassword("WrongPassword")
	assert.False(t, isInvalid)

	deleteAllAccounts()
}

func TestGenerateTokenConfirmTel(t *testing.T) {

	service := newAccountService(appService)
	userAccount, _ := NewUserAccount()
	userAccount.SetTel("+123121234567")

	token, err := service.GenerateTokenConfirmTel("+123121234567", userAccount)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	deleteAllAccounts()
}

func TestValidateTokenConfirmTel(t *testing.T) {

	service := newAccountService(appService)
	userAccount, _ := NewUserAccount()
	userAccount.SetTel("+123121234567")

	token, _ := service.GenerateTokenConfirmTel("+123121234567", userAccount)

	ok, err := service.ValidateTokenConfirmTel(token, "+123121234567", userAccount)
	assert.NoError(t, err)
	assert.True(t, ok)

	deleteAllAccounts()
}

func TestGeneratePasscodeConfirmTel(t *testing.T) {

	service := newAccountService(appService)
	userAccount, _ := NewUserAccount()
	userAccount.SetTel("+123121234567")

	passcode, err := service.GeneratePasscodeConfirmTel("+123121234567", userAccount)
	assert.NoError(t, err)
	assert.Len(t, passcode, 8)

	deleteAllAccounts()
}

func TestValidatePasscodeConfirmTel(t *testing.T) {

	service := newAccountService(appService)
	userAccount, _ := NewUserAccount()
	userAccount.SetTel("+123121234567")

	passcode, _ := service.GeneratePasscodeConfirmTel("+123121234567", userAccount)

	ok, err := service.ValidatePasscodeConfirmTel(passcode, "+123121234567", userAccount)
	assert.NoError(t, err)
	assert.True(t, ok)

	deleteAllAccounts()
}

func TestGenerateTokenConfirmEmail(t *testing.T) {

	service := newAccountService(appService)
	userAccount, _ := NewUserAccount()
	userAccount.SetEmail("user@example.com")

	token, err := service.GenerateTokenConfirmEmail("user@example.com", userAccount)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	deleteAllAccounts()
}

func TestValidateTokenConfirmEmail(t *testing.T) {

	service := newAccountService(appService)
	userAccount, _ := NewUserAccount()
	userAccount.SetEmail("user@example.com")

	token, _ := service.GenerateTokenConfirmEmail("user@example.com", userAccount)

	ok, err := service.ValidateTokenConfirmEmail(token, "user@example.com", userAccount)
	assert.NoError(t, err)
	assert.True(t, ok)

	deleteAllAccounts()
}

func TestGeneratePasscodeConfirmEmail(t *testing.T) {

	service := newAccountService(appService)
	userAccount, _ := NewUserAccount()
	userAccount.SetEmail("user@example.com")

	passcode, err := service.GeneratePasscodeConfirmEmail("user@example.com", userAccount)
	assert.NoError(t, err)
	assert.Len(t, passcode, 8)

	deleteAllAccounts()
}

func TestValidatePasscodeConfirmEmail(t *testing.T) {

	service := newAccountService(appService)
	userAccount, _ := NewUserAccount()
	userAccount.SetEmail("user@example.com")

	passcode, _ := service.GeneratePasscodeConfirmEmail("user@example.com", userAccount)

	ok, err := service.ValidatePasscodeConfirmEmail(passcode, "user@example.com", userAccount)
	assert.NoError(t, err)
	assert.True(t, ok)

	deleteAllAccounts()
}

func TestCreateUserAccount(t *testing.T) {

	service := newAccountService(appService)
	userAccount, _ := NewUserAccount()
	userAccount.SetUsername("testuser")
	userAccount.SetTel("+123121234567")
	userAccount.SetEmail("user@example.com")
	userAccount.SetPassword("StrongPass1")

	err := service.CreateUserAccount(userAccount)
	assert.NoError(t, err)

	retrievedAccount, err := service.FindByID(userAccount.ID)
	assert.NoError(t, err)
	assert.Equal(t, userAccount.Username, retrievedAccount.Username)

	deleteAllAccounts()
}

func TestUpdateUserAccount(t *testing.T) {

	service := newAccountService(appService)
	userAccount, _ := NewUserAccount()
	userAccount.SetUsername("testuser")
	userAccount.SetTel("+123121234567")
	userAccount.SetEmail("user@example.com")
	userAccount.SetPassword("StrongPass1")

	err := service.CreateUserAccount(userAccount)
	assert.NoError(t, err)

	userAccount.SetUsername("updateduser")
	err = service.UpdateUserAccount(userAccount)
	assert.NoError(t, err)

	retrievedAccount, err := service.FindByID(userAccount.ID)
	assert.NoError(t, err)
	assert.Equal(t, "updateduser", retrievedAccount.Username)

	deleteAllAccounts()
}

func TestFindByID(t *testing.T) {

	service := newAccountService(appService)
	userAccount, _ := NewUserAccount()
	userAccount.SetUsername("testuser")
	userAccount.SetTel("+123121234567")
	userAccount.SetEmail("user@example.com")
	userAccount.SetPassword("StrongPass1")

	err := service.CreateUserAccount(userAccount)
	assert.NoError(t, err)

	retrievedAccount, err := service.FindByID(userAccount.ID)
	assert.NoError(t, err)
	assert.Equal(t, userAccount.Username, retrievedAccount.Username)

	deleteAllAccounts()
}

func TestFindByUsername(t *testing.T) {

	service := newAccountService(appService)
	userAccount, _ := NewUserAccount()
	userAccount.SetUsername("testuser")
	userAccount.SetTel("+123121234567")
	userAccount.SetEmail("user@example.com")
	userAccount.SetPassword("StrongPass1")

	err := service.CreateUserAccount(userAccount)
	assert.NoError(t, err)

	retrievedAccount, err := service.FindByUsername("testuser")
	assert.NoError(t, err)
	assert.Equal(t, userAccount.Username, retrievedAccount.Username)

	deleteAllAccounts()
}

func TestFindByTel(t *testing.T) {

	service := newAccountService(appService)
	userAccount, _ := NewUserAccount()
	userAccount.SetUsername("testuser")
	userAccount.SetTel("+123121234567")
	userAccount.SetEmail("user@example.com")
	userAccount.SetPassword("StrongPass1")

	err := service.CreateUserAccount(userAccount)
	assert.NoError(t, err)

	retrievedAccount, err := service.FindByTel("+123121234567")
	assert.NoError(t, err)
	assert.Equal(t, userAccount.Tel, retrievedAccount.Tel)

	deleteAllAccounts()
}

func TestFindByNormalizedEmail(t *testing.T) {

	service := newAccountService(appService)
	userAccount, _ := NewUserAccount()
	userAccount.SetUsername("testuser")
	userAccount.SetTel("+123121234567")
	userAccount.SetEmail("user@example.com")
	userAccount.SetPassword("StrongPass1")

	err := service.CreateUserAccount(userAccount)
	assert.NoError(t, err)

	retrievedAccount, err := service.FindByNormalizedEmail("USER@EXAMPLE.COM")
	assert.NoError(t, err)
	assert.Equal(t, userAccount.NormalizedEmail, retrievedAccount.NormalizedEmail)

	deleteAllAccounts()
}
