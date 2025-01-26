package messenger

import (
	"go-auth/internal/config"
	"go-auth/internal/util/utilhttp"
	xlog "go-auth/internal/util/utillog"
	"strings"
)

type AppMessenger interface {
	// SendSms(text string, tel string)
	// SendEmail(html string, subject string, email string)

	SendPasscodeToTel(code string, tel string, lang string)
	SendPasscodeToEmail(code string, email string, lang string)
}

type defaultAppMessenger struct {
	Debug  bool
	config config.AppConfigMessenger
	// logger logger.AppLogger
}

func NewAppMessenger(config *config.AppConfig,

// logger logger.AppLogger
) (res AppMessenger) {

	// queue
	// background task
	// http client || tmp file
	// config
	// logger // TODO named logger

	res = &defaultAppMessenger{
		Debug:  config.Debug,
		config: config.Messenger,
		// logger: logger,
	}

	return
}

func (x *defaultAppMessenger) SendPasscodeToTel(passcode string, tel string, lang string) {
	serviceCode := "sms-passcode" // _ -

	formValues := map[string]string{
		"to":       tel,
		"passcode": passcode,
		"lang":     lang,
	}

	if x.Debug || x.config.Stdout {
		xlog.Info(serviceCode, formValues)
	}

	URL := strings.ReplaceAll(x.config.ServiceURL, "{code}", serviceCode)

	_, err := utilhttp.PostFormURL(URL, nil, nil, formValues)

	if err != nil {
		xlog.Error("error from sms service: %v", err)
	}

}

func (x *defaultAppMessenger) SendPasscodeToEmail(passcode string, email string, lang string) {

	serviceCode := "email-passcode" // _ -

	formValues := map[string]string{
		"to":       email,
		"passcode": passcode,
		"lang":     lang,
	}

	if x.Debug || x.config.Stdout {
		xlog.Info(serviceCode, formValues)
	}

	URL := strings.ReplaceAll(x.config.ServiceURL, "{code}", serviceCode)

	_, err := utilhttp.PostFormURL(URL, nil, nil, formValues)

	if err != nil {
		xlog.Error("error from sms service: %v", err)
	}

}
