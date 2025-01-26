package mvc

import (
	utilstring "go-auth/internal/util/utilstring"
	"slices" // Ensure the import path is correct or replace with appropriate package
	"strconv"
	"unicode"
)

// ModelBaseDTO is a base struct for models or DTOs with validation error handling.
type ModelBaseDTO struct {
	Errors []ErrorMessage `json:"errors,omitempty"`
}

// AddError adds an error message to the model.
func (x *ModelBaseDTO) AddError(code string, msg string) {
	x.Errors = append(x.Errors, ErrorMessage{Code: code, Message: msg})
}

// RemoveError removes an error message from the model by its code.
func (x *ModelBaseDTO) RemoveError(code string) {
	x.Errors = slices.DeleteFunc(x.Errors, func(x ErrorMessage) bool {
		return x.Code == code
	})
}

// IsModelValid checks if the model has any validation errors.
func (x *ModelBaseDTO) IsModelValid() bool {
	return len(x.Errors) == 0
}

// ModelValidatorStr assists in validating fields in ModelBaseDTO.
type ModelValidatorStr struct {
	model      *ModelBaseDTO
	lang       UserLang
	fieldName  string
	fieldValue string
	fieldTitle string
	hasError   bool
}

// NewModelValidatorStr creates a new ModelValidator for a specific field.
func (x *ModelBaseDTO) NewModelValidatorStr(lang UserLang, fieldName string, fieldTitle string, fieldValue string, maxLen int) *ModelValidatorStr {
	res := &ModelValidatorStr{
		model:      x,
		lang:       lang,
		fieldName:  fieldName,
		fieldValue: fieldValue,
		fieldTitle: fieldTitle,
		hasError:   false,
	}

	res.LengthRange(0, maxLen)

	return res
}

// LengthRange checks if the length of the field's value is within the min and max limits.
func (x *ModelValidatorStr) LengthRange(minLen int, maxLen int) (hasError bool) {
	v := x.fieldValue
	if minLen > 0 && len(v) < minLen {
		x.model.AddError(x.fieldName,
			x.lang.Lang("The '{0}' must be at least {1} characters.", /*Lang*/
				x.lang.Lang(x.fieldTitle),
				strconv.Itoa(minLen)))
		return true
	}

	if maxLen > 0 && len(v) > maxLen {
		x.model.AddError(x.fieldName,
			x.lang.Lang("The '{0}' must be at most {1} characters.", /*Lang*/
				x.lang.Lang(x.fieldTitle),
				strconv.Itoa(maxLen)))
		return true
	}

	return false
}

// LengthMax checks if the length of the field's value is within the max limit.
func (x *ModelValidatorStr) LengthMax(maxLen int) (hasError bool) {
	return x.LengthRange(0, maxLen)
}

// Required checks if the field's value is not empty.
func (x *ModelValidatorStr) Required() (hasError bool) {
	v := x.fieldValue
	if len(v) < 1 {

		x.model.AddError(x.fieldName,
			x.lang.Lang("Field '{0}' is required.", /*Lang*/
				x.lang.Lang(x.fieldTitle),
			))

		return true
	}
	return false
}

// Password checks if the field's value meets the password complexity requirements.
// minLen=8 && (a-z && A-Z && 0-9)
func (x *ModelValidatorStr) Password(minLen int) (hasError bool) {
	v := x.fieldValue

	if minLen > 0 && x.LengthRange(minLen, 0) {
		return true
	}

	if !hasDigitLowerUpper(v) {

		t := x.lang.Lang(x.fieldTitle)
		x.model.AddError(x.fieldName, x.lang.Lang("The '{0}' must have at least one digit ('0'-'9')." /*Lang*/, t))
		x.model.AddError(x.fieldName, x.lang.Lang("The '{0}' must have at least one lowercase letter ('a'-'z')." /*Lang*/, t))
		x.model.AddError(x.fieldName, x.lang.Lang("The '{0}' must have at least one uppercase letter ('A'-'Z')." /*Lang*/, t))

		return true
	}

	return false
}

// Keyword checks if the field's value match keyword.
func (x *ModelValidatorStr) Keyword(keyword string) (hasError bool) {
	v := x.fieldValue
	if v != keyword {

		x.model.AddError(x.fieldName,
			x.lang.Lang("Keyword '{0}' is required.", /*Lang*/
				keyword,
			))

		return true
	}
	return false
}

// hasDigitLowerUpper checks if the string contains at least one digit, one lowercase, and one uppercase character.
func hasDigitLowerUpper(s string) (hasError bool) {
	hasDigit := false
	hasLower := false
	hasUpper := false

	for _, char := range s {
		switch {
		case unicode.IsDigit(char):
			hasDigit = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsUpper(char):
			hasUpper = true
		}
		// Early exit if all conditions are met
		if hasDigit && hasLower && hasUpper {
			return true
		}
	}

	return hasDigit && hasLower && hasUpper
}

// Email checks if the field's value is a valid email.
// minLen=6  a@a.aa
func (x *ModelValidatorStr) Email(minLen int) (hasError bool) {
	v := x.fieldValue

	if minLen > 0 && x.LengthRange(minLen, 0) {
		return true
	}

	if len(v) > 0 && !utilstring.IsEmail(v) {
		x.model.AddError(x.fieldName,
			x.lang.Lang("Email '{0}' is invalid." /*Lang*/, v),
		)

		return true
	}

	return false
}

// Digits checks if the field's value is [0-9]+.
func (x *ModelValidatorStr) Digits() (hasError bool) {
	v := x.fieldValue

	if len(v) > 0 && !utilstring.IsDigits(v) {
		t := x.lang.Lang(x.fieldTitle)
		x.model.AddError(x.fieldName, x.lang.Lang("The '{0}' must contain only digits ('0'-'9')." /*Lang*/, t))
		return true
	}

	return false
}

// // TelBody checks if the field's value is a valid phone number body.
// func (x *ModelValidatorStr) TelBody() (hasError bool) {
// 	v := x.fieldValue
// 	if len(v) > 0 && !utilstring.IsTelBody(v) {
// 		x.model.AddError(utileldName,
// 			x.lang.Lang("Please enter a valid phone number" /*Lang*/),
// 		)

// 		return true
// 	}
// 	return false

// }

// // TelPrefix checks if the field's value is a valid phone number prefix.
// func (x *ModelValidatorStr) TelPrefix() (hasError bool) {
// 	v := x.fieldValue
// 	if len(v) > 0 && !utilstring.IsTelPrefix(v) {
// 		x.model.AddError(x.fieldName,
// 			x.lang.Lang("Please enter a valid phone number" /*Lang*/),
// 		)

// 		return true
// 	}
// 	return false

// }

// Tel checks if the field's value is a valid phone number.
func (x *ModelValidatorStr) Tel() (hasError bool) {
	v := x.fieldValue
	if len(v) > 0 && !utilstring.IsTelFull(v) {
		x.model.AddError(x.fieldName,
			x.lang.Lang("Please enter a valid phone number." /*Lang*/),
		)

		return true
	}
	return false

}
