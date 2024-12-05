package utilstring

import (
	"net/url"
	"regexp"
	"strings"
)

func NormalizeText(t string) string {
	return strings.ToLower(strings.TrimSpace(t))
}

func NormalizePhoneNumber(phoneNumber string) string {

	//	phoneNumber = strings.TrimSpace(phoneNumber)

	re := regexp.MustCompile(`[^0-9+]`)
	phoneNumber = re.ReplaceAllString(phoneNumber, "")

	return phoneNumber
}

func NormalizeEmail(email string) string {

	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return email
	}

	localPart := parts[0]
	domainPart := parts[1]

	// Find the '+' character in the local part
	if plusIndex := strings.Index(localPart, "+"); plusIndex != -1 {
		// If found, remove everything after the '+'
		localPart = localPart[:plusIndex]
	}

	// Return the normalized email
	return NormalizeText(localPart + "@" + domainPart)
}

// // IsPhoneNumberPrefix country prefix +123
// func IsPhoneNumberPrefix(str string) bool {
// 	// Compiles the regular expression and checks if the string matches
// 	return regexp.MustCompile(`^[+][0-9]{1,3}$`).MatchString(str)
// }

// // IsPhoneNumberBody number body part (without prefix)
// func IsPhoneNumberBody(str string) bool {
// 	// Compiles the regular expression and checks if the string matches
// 	return regexp.MustCompile(`^[0-9]{7,12}$`).MatchString(str)
// }

// IsPhoneNumberFull full number
func IsPhoneNumberFull(str string) bool {
	// Compiles the regular expression and checks if the string matches
	return regexp.MustCompile(`^[+][0-9]{9,18}$`).MatchString(str)
}

// IsEmail full number
func IsEmail(str string) bool {

	return strings.Contains(str, "@") || strings.Contains(str, ".")
}

// AppendURL join like path?args[0]=args[1]&args[2]=args[3]#args[4]
// ignore query key-value pair or fragment if is empty
func AppendURL(path string, args ...string) string {

	count := len(args)
	pairs := count / 2

	if pairs > 0 {

		u := url.URL{
			Path: path,
		}
		query := u.Query()
		for i := 0; i < pairs; i++ {
			k := args[i*2]
			v := args[i*2+1]
			if k != "" && v != "" {
				query.Add(k, v) // this not keep order, internal sort by key on encode
			}

		}

		u.RawQuery = query.Encode()

		if count%2 == 1 {
			v := args[count-1]
			if v != "" {
				u.Fragment = args[count-1]
			}
		}

		return u.String()

	}

	return path
}
