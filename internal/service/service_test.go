package service

var appService AppService

func beginTest() {
	if appService == nil {
		appService = MustNewAppServiceTesting()
	}
}

func endTest() {

	appService.Repository().Where("1=1").Delete(&UserAccount{})
}
