package messenger

import (
	"go-auth/internal/config"
	"go-auth/internal/util/utilhttp"
	xlog "go-auth/internal/util/utillog"
	"strings"
)

type AppMessenger interface {
	// SendSms(text string, phoneNumber string)
	// SendEmail(html string, subject string, email string)

	SendSecretCodeToPhoneNumber(code string, phoneNumber string, lang string)
	SendSecretCodeToEmail(code string, email string, lang string)
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

// func (x *appMessenger) SendSms(text string, phoneNumber string) {

// 	data := map[string]string{
// 		"type":         "sms",
// 		"phone_number": phoneNumber,
// 		"text":         text,
// 	}

// 	if x.Debug {

// 		x.logger.ZapLogger().Debugw("SendSms", "type", "sms", "phone_number", phoneNumber, "text", text)
// 	}

// 	_, err := utilhttp.Post(x.config.PhoneNumberURL, data, nil)

// 	if err != nil {
// 		x.logger.ZapLogger().Errorf("Error from sms service: %v", err)
// 	}
// }

func (x *defaultAppMessenger) SendSecretCodeToPhoneNumber(secretCode string, phoneNumber string, lang string) {
	serviceCode := "sms-secret-code" // _ -

	formValues := map[string]string{
		"to":          phoneNumber,
		"secret_code": secretCode,
		"lang":        lang,
	}

	if x.Debug || x.config.Stdout {
		xlog.Info(serviceCode, formValues)
	}

	URL := strings.ReplaceAll(x.config.ServiceURL, "{code}", serviceCode)

	_, err := utilhttp.PostFormURL(URL, nil, formValues, nil)

	if err != nil {
		xlog.Error("Error from sms service: %v", err)
	}

}

func (x *defaultAppMessenger) SendSecretCodeToEmail(secretCode string, email string, lang string) {

	serviceCode := "email-secret-code" // _ -

	formValues := map[string]string{
		"to":          email,
		"secret_code": secretCode,
		"lang":        lang,
	}

	if x.Debug || x.config.Stdout {
		xlog.Info(serviceCode, formValues)
	}

	URL := strings.ReplaceAll(x.config.ServiceURL, "{code}", serviceCode)

	_, err := utilhttp.PostFormURL(URL, nil, formValues, nil)

	if err != nil {
		xlog.Error("Error from sms service: %v", err)
	}

}
