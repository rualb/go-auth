package token

/*
Practical Recommendations (source: AI)
Standard Practice: For HS256, a key length of 256 bits (32 bytes) is typically used and recommended.
This length balances security and performance effectively.
Extended Length: If you want to use a longer key, such as 512 bits (64 bytes),
it is acceptable but may not provide significant additional security benefits over a 256-bit key.
*/

/*
Rfc 6238
HMACSHA1 blocksize 64 output 20
Base32
*/

import (
	"encoding/base32"
	"fmt"
	"time"

	otp "github.com/pquerna/otp"
	totp "github.com/pquerna/otp/totp"

	"crypto/hmac"
	"crypto/sha256"
)

var b32NoPadding = base32.StdEncoding.WithPadding(base32.NoPadding)

const (
	PasscodeLenDefault       = 8                   // digits
	PasscodeLifetimeDefault  = uint(30)            // seconds
	PasscodeAlgorithmDefault = otp.AlgorithmSHA256 // AlgorithmSHA512 AlgorithmSHA1
	SecretKeyMinLen          = 32                  // sha256.BlockSize
)

type ConfigTotp struct {
	Scope     string // e.g., signup, password_reset
	Digits    int    // Passcode length
	SecretKey []byte // Must be >= SecretKeyMinLen Base64
	// SecretKeyExt string // Extra key from user accaunt SecurityStamp
	Time int // For debug/test mode
}

func (x ConfigTotp) Validate() error {
	if len(x.SecretKey) < SecretKeyMinLen {
		return fmt.Errorf("parameter SecretKey must be at least %d characters long", SecretKeyMinLen)
	}
	if x.Scope == "" {
		return fmt.Errorf("parameter Scope cannot be empty")
	}
	return nil
}
func NewConfigTotp(scope string, secretKey []byte) ConfigTotp {

	return ConfigTotp{
		Scope:     scope,
		Digits:    PasscodeLenDefault,
		SecretKey: secretKey,

		Time: 0,
	}
}

func convertToSecretBase32(config ConfigTotp) (strBase32Encoded string) {

	// Create a new HMAC using SHA-256
	h := hmac.New(sha256.New, config.SecretKey)
	// Write the message to the HMAC object
	h.Write([]byte(config.Scope))
	// Get the final HMAC hash
	hash := h.Sum(nil)

	// b32NoPadding
	strBase32Encoded = b32NoPadding.EncodeToString(hash)

	return strBase32Encoded

}

func convertConfig(config ConfigTotp) totp.ValidateOpts {
	return totp.ValidateOpts{
		Digits:    otp.Digits(config.Digits),
		Period:    PasscodeLifetimeDefault,
		Algorithm: PasscodeAlgorithmDefault,
		Skew:      1,
	}
}

func convertTime(config ConfigTotp) time.Time {
	t := time.Now().UTC() // now

	// if config.Time not empty use it as time source (for debug and test)
	if config.Time != 0 {
		t = time.Unix(int64(config.Time), 0)
	}

	return t
}

// GeneratePasscode generates a TOTP passcode based on info, scope, and security stamp.
func GeneratePasscode(config ConfigTotp) (string, error) {

	if err := config.Validate(); err != nil {
		return "", err
	}

	// create secret from any string by master key
	secretBase32 := convertToSecretBase32(config)

	// if len(key) > blocksize  If key is too big, hash it.

	return totp.GenerateCodeCustom(secretBase32 /*base32*/, convertTime(config), convertConfig(config))
}

// ValidatePasscode validates if the provided passcode matches the expected value.
func ValidatePasscode(passcode string, config ConfigTotp) (bool, error) {

	if err := config.Validate(); err != nil {
		return false, err
	}

	// create secret from any string by master key
	secretBase32 := convertToSecretBase32(config)

	return totp.ValidateCustom(passcode, secretBase32 /*base32*/, convertTime(config), convertConfig(config))
}
