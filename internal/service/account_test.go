package service

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewUserAccount(t *testing.T) {
	beginTest()
	defer endTest()
	account, err := NewUserAccount()
	assert.NoError(t, err)
	assert.NotNil(t, account)
	assert.NotEmpty(t, account.ID)
	assert.True(t, account.CreatedAt.Before(time.Now().Add(time.Millisecond)))

}

func TestSetUsername(t *testing.T) {
	beginTest()
	defer endTest()
	account, _ := NewUserAccount()
	account.SetUsername("TestUser")
	assert.Equal(t, "testuser", account.Username)
}

func TestSetPhoneNumber(t *testing.T) {
	beginTest()
	defer endTest()
	account, _ := NewUserAccount()
	account.SetPhoneNumber("+123121234567")
	assert.Equal(t, "+123121234567", account.PhoneNumber)
}

func TestSetEmail(t *testing.T) {
	beginTest()
	defer endTest()
	account, _ := NewUserAccount()
	account.SetEmail("User@Example.com")
	assert.Equal(t, "User@Example.com", account.Email)
	assert.Equal(t, "user@example.com", account.NormalizedEmail)
}

func TestSetPassword(t *testing.T) {
	beginTest()
	defer endTest()
	account, _ := NewUserAccount()
	err := account.SetPassword("StrongPass1")
	assert.NoError(t, err)
	assert.NotEmpty(t, account.PasswordHash)

	isValid := account.CompareHashAndPassword("StrongPass1")
	assert.True(t, isValid)

	isInvalid := account.CompareHashAndPassword("WrongPassword")
	assert.False(t, isInvalid)
}

func TestGenerateTokenConfirmPhoneNumber(t *testing.T) {
	beginTest()
	defer endTest()
	service := newAccountService(appService)
	userAccount, _ := NewUserAccount()
	userAccount.SetPhoneNumber("+123121234567")

	token, err := service.GenerateTokenConfirmPhoneNumber("+123121234567", userAccount)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestValidateTokenConfirmPhoneNumber(t *testing.T) {
	beginTest()
	defer endTest()
	service := newAccountService(appService)
	userAccount, _ := NewUserAccount()
	userAccount.SetPhoneNumber("+123121234567")

	token, _ := service.GenerateTokenConfirmPhoneNumber("+123121234567", userAccount)

	ok, err := service.ValidateTokenConfirmPhoneNumber(token, "+123121234567", userAccount)
	assert.NoError(t, err)
	assert.True(t, ok)
}

func TestGeneratePasscodeConfirmPhoneNumber(t *testing.T) {
	beginTest()
	defer endTest()
	service := newAccountService(appService)
	userAccount, _ := NewUserAccount()
	userAccount.SetPhoneNumber("+123121234567")

	passcode, err := service.GeneratePasscodeConfirmPhoneNumber("+123121234567", userAccount)
	assert.NoError(t, err)
	assert.Len(t, passcode, 8)
}

func TestValidatePasscodeConfirmPhoneNumber(t *testing.T) {
	beginTest()
	defer endTest()
	service := newAccountService(appService)
	userAccount, _ := NewUserAccount()
	userAccount.SetPhoneNumber("+123121234567")

	passcode, _ := service.GeneratePasscodeConfirmPhoneNumber("+123121234567", userAccount)

	ok, err := service.ValidatePasscodeConfirmPhoneNumber(passcode, "+123121234567", userAccount)
	assert.NoError(t, err)
	assert.True(t, ok)
}

func TestGenerateTokenConfirmEmail(t *testing.T) {
	beginTest()
	defer endTest()
	service := newAccountService(appService)
	userAccount, _ := NewUserAccount()
	userAccount.SetEmail("user@example.com")

	token, err := service.GenerateTokenConfirmEmail("user@example.com", userAccount)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestValidateTokenConfirmEmail(t *testing.T) {
	beginTest()
	defer endTest()
	service := newAccountService(appService)
	userAccount, _ := NewUserAccount()
	userAccount.SetEmail("user@example.com")

	token, _ := service.GenerateTokenConfirmEmail("user@example.com", userAccount)

	ok, err := service.ValidateTokenConfirmEmail(token, "user@example.com", userAccount)
	assert.NoError(t, err)
	assert.True(t, ok)
}

func TestGeneratePasscodeConfirmEmail(t *testing.T) {
	beginTest()
	defer endTest()
	service := newAccountService(appService)
	userAccount, _ := NewUserAccount()
	userAccount.SetEmail("user@example.com")

	passcode, err := service.GeneratePasscodeConfirmEmail("user@example.com", userAccount)
	assert.NoError(t, err)
	assert.Len(t, passcode, 8)
}

func TestValidatePasscodeConfirmEmail(t *testing.T) {
	beginTest()
	defer endTest()
	service := newAccountService(appService)
	userAccount, _ := NewUserAccount()
	userAccount.SetEmail("user@example.com")

	passcode, _ := service.GeneratePasscodeConfirmEmail("user@example.com", userAccount)

	ok, err := service.ValidatePasscodeConfirmEmail(passcode, "user@example.com", userAccount)
	assert.NoError(t, err)
	assert.True(t, ok)
}

func TestCreateUserAccount(t *testing.T) {
	beginTest()
	defer endTest()
	service := newAccountService(appService)
	userAccount, _ := NewUserAccount()
	userAccount.SetUsername("testuser")
	userAccount.SetPhoneNumber("+123121234567")
	userAccount.SetEmail("user@example.com")
	userAccount.SetPassword("StrongPass1")

	err := service.CreateUserAccount(userAccount)
	assert.NoError(t, err)

	retrievedAccount, err := service.FindByID(userAccount.ID)
	assert.NoError(t, err)
	assert.Equal(t, userAccount.Username, retrievedAccount.Username)
}

func TestUpdateUserAccount(t *testing.T) {
	beginTest()
	defer endTest()
	service := newAccountService(appService)
	userAccount, _ := NewUserAccount()
	userAccount.SetUsername("testuser")
	userAccount.SetPhoneNumber("+123121234567")
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
}

func TestFindByID(t *testing.T) {
	beginTest()
	defer endTest()
	service := newAccountService(appService)
	userAccount, _ := NewUserAccount()
	userAccount.SetUsername("testuser")
	userAccount.SetPhoneNumber("+123121234567")
	userAccount.SetEmail("user@example.com")
	userAccount.SetPassword("StrongPass1")

	err := service.CreateUserAccount(userAccount)
	assert.NoError(t, err)

	retrievedAccount, err := service.FindByID(userAccount.ID)
	assert.NoError(t, err)
	assert.Equal(t, userAccount.Username, retrievedAccount.Username)
}

func TestFindByUsername(t *testing.T) {
	beginTest()
	defer endTest()
	service := newAccountService(appService)
	userAccount, _ := NewUserAccount()
	userAccount.SetUsername("testuser")
	userAccount.SetPhoneNumber("+123121234567")
	userAccount.SetEmail("user@example.com")
	userAccount.SetPassword("StrongPass1")

	err := service.CreateUserAccount(userAccount)
	assert.NoError(t, err)

	retrievedAccount, err := service.FindByUsername("testuser")
	assert.NoError(t, err)
	assert.Equal(t, userAccount.Username, retrievedAccount.Username)
}

func TestFindByPhoneNumber(t *testing.T) {
	beginTest()
	defer endTest()
	service := newAccountService(appService)
	userAccount, _ := NewUserAccount()
	userAccount.SetUsername("testuser")
	userAccount.SetPhoneNumber("+123121234567")
	userAccount.SetEmail("user@example.com")
	userAccount.SetPassword("StrongPass1")

	err := service.CreateUserAccount(userAccount)
	assert.NoError(t, err)

	retrievedAccount, err := service.FindByPhoneNumber("+123121234567")
	assert.NoError(t, err)
	assert.Equal(t, userAccount.PhoneNumber, retrievedAccount.PhoneNumber)
}

func TestFindByNormalizedEmail(t *testing.T) {
	beginTest()
	defer endTest()
	service := newAccountService(appService)
	userAccount, _ := NewUserAccount()
	userAccount.SetUsername("testuser")
	userAccount.SetPhoneNumber("+123121234567")
	userAccount.SetEmail("user@example.com")
	userAccount.SetPassword("StrongPass1")

	err := service.CreateUserAccount(userAccount)
	assert.NoError(t, err)

	retrievedAccount, err := service.FindByNormalizedEmail("USER@EXAMPLE.COM")
	assert.NoError(t, err)
	assert.Equal(t, userAccount.NormalizedEmail, retrievedAccount.NormalizedEmail)
}
