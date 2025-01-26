package service

import (
	"os"
	"testing"
)

var appService AppService

func TestMain(m *testing.M) {
	beginTest()
	code := m.Run()
	deleteAllAccounts()
	os.Exit(code)
}

func beginTest() {
	if appService == nil {
		appService = MustNewAppServiceTesting()
	}

	deleteAllAccounts()
}

func deleteAllAccounts() {

	appService.Repository().Where("1=1").Delete(&UserAccount{})
}
