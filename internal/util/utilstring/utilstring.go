package utilstring

import (
	"regexp"
	"strings"
	"unicode"
)

func NormalizeText(t string) string {
	return strings.ToLower(strings.TrimSpace(t))
}

func NormalizeTel(tel string) string {

	//	tel = strings.TrimSpace(tel)

	re := regexp.MustCompile(`[^0-9+]`)
	tel = re.ReplaceAllString(tel, "")

	return tel
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

// // IsTelPrefix country prefix +123
// func IsTelPrefix(str string) bool {
// 	// Compiles the regular expression and checks if the string matches
// 	return regexp.MustCompile(`^[+][0-9]{1,3}$`).MatchString(str)
// }

// // IsTelBody number body part (without prefix)
// func IsTelBody(str string) bool {
// 	// Compiles the regular expression and checks if the string matches
// 	return regexp.MustCompile(`^[0-9]{7,12}$`).MatchString(str)
// }

// IsTelFull full number
func IsTelFull(str string) bool {
	// Compiles the regular expression and checks if the string matches
	return regexp.MustCompile(`^[+][0-9]{9,18}$`).MatchString(str)
}

// IsEmail full number
func IsEmail(str string) bool {

	return strings.Contains(str, "@") || strings.Contains(str, ".")
}

// IsDigits numeric [0-9]+
func IsDigits(str string) bool {

	if str == "" {
		return false // Consider an empty string as non-numeric.
	}
	for _, r := range str {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true

}
