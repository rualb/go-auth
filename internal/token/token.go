package token

import (
	"fmt"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
)

var (
	JwtSigningMethodDefault = jwt.SigningMethodHS256 // sha512

)

const (
	ScopeAuth   = "auth"   // real user
	ScopeSignup = "signup" // user signup (registration) process
)

type TokenPersist interface {
	AuthTokenClaims() *TokenClaimsDTO
	CreateAuthTokenWithClaims(claims *TokenClaimsDTO) error
	DeleteAuthToken()
	RotateAuthToken(forceRotate bool)
}

type TokenClaimsDTO struct {
	PhoneNumber string           `json:"phone_number,omitempty"`
	UserID      string           `json:"user_id,omitempty"`
	Email       string           `json:"email,omitempty"`
	Scope       jwt.ClaimStrings `json:"scope,omitempty"` // as []string and as string
	jwt.RegisteredClaims
}

// SetLifetime maxAge is seconds
func (x *TokenClaimsDTO) SetLifetime(maxAge time.Duration) {
	lifetime := maxAge
	now := time.Now().UTC() // now

	iss := now
	exp := now.Add(lifetime)
	x.IssuedAt = jwt.NewNumericDate(iss)
	x.ExpiresAt = jwt.NewNumericDate(exp)
}
func (x *TokenClaimsDTO) IsIssuedBy(issuer string) bool {

	return x.Issuer == issuer
}

func (x *TokenClaimsDTO) SetIssuer(issuer string) {

	x.Issuer = issuer
}

func (x *TokenClaimsDTO) AddScope(scope string) {

	x.Scope = append(x.Scope, scope)
}

func (x *TokenClaimsDTO) HasScope(scope string) bool {

	if len(x.Scope) > 0 {

		for _, x := range x.Scope {
			if x == scope {
				return true
			}
		}
	}

	return false
}

func (x TokenClaimsDTO) IsPhoneNumberMatch(value string) bool {

	return x.PhoneNumber == value
}

// func (x TokenClaimsDTO) IsAudienceMatch(value string) bool {

// 	return len(x.Audience) > 0 && slices.Contains(x.Audience, value)
// }

func (x TokenClaimsDTO) IsEmpty() bool {

	return x.IssuedAt == nil || x.ExpiresAt == nil
}
func (x TokenClaimsDTO) IsValid() bool {
	now := time.Now().UTC().Unix() // now

	return !x.IsEmpty() && x.IssuedAt.Unix() <= now && now <= x.ExpiresAt.Unix()
}

func (x TokenClaimsDTO) IsSignedIn() bool {

	return x.IsValid() && x.UserID != ""
}

func (x TokenClaimsDTO) NeedRotation() bool {
	now := time.Now().UTC().Unix() // now

	return (x.ExpiresAt.Unix() - x.IssuedAt.Unix()) >= now
}
func (x TokenClaimsDTO) Rotate(forceRotate bool) *TokenClaimsDTO {

	if x.IsValid() && (forceRotate || x.NeedRotation()) {

		claimsNew := x // create a copy

		// config := x.appService.Config()

		lifetime := claimsNew.ExpiresAt.Time.Sub(claimsNew.IssuedAt.Time)
		lifetime = min(lifetime, time.Hour*24*30) // max 30 days

		claimsNew.SetLifetime(lifetime) // extend lifetime

		// rotate token

		return &claimsNew

	}

	return nil

}

type SecretSourceCurrent interface {
	CurrentKey() (id string, secret []byte, err error)
}

type SecretSourceByID interface {
	KeyByID(id string) (secret []byte, err error)
}

func CreateToken(claims *TokenClaimsDTO, secretSourceAuth SecretSourceCurrent) (string, error) {

	// "key is of invalid type" Key needs to be a []byte

	secretID, secret, err := secretSourceAuth.CurrentKey()
	if err != nil {
		return "", err
	}
	// sha512
	token := jwt.NewWithClaims(JwtSigningMethodDefault, claims)
	token.Header["kid"] = secretID

	tokenString, err := token.SignedString(secret)
	if err != nil {
		return "", fmt.Errorf("error on creating jwt token: %v", err)
	}

	return tokenString, nil
}

func ParseToken(tokenString string, secretSourceByID SecretSourceByID) (*TokenClaimsDTO, error) {

	token, err := jwt.ParseWithClaims(tokenString, new(TokenClaimsDTO), func(token *jwt.Token) (interface{}, error) {

		key, err := JwtSecretSearch(token, secretSourceByID)

		return key, err
	})

	if err != nil {
		return nil, err
	}

	if token != nil && token.Valid {
		claims, _ := token.Claims.(*TokenClaimsDTO)

		if claims != nil && claims.IsValid() {
			return claims, nil
		}
	}

	return nil, nil

}

func JwtSecretSearch(token *jwt.Token, secretSourceByID SecretSourceByID) (interface{}, error) {

	if token == nil {
		return nil, fmt.Errorf("jwt token is null")
	}

	if kid, ok := token.Header["kid"]; ok {

		secret, err := secretSourceByID.KeyByID(kid.(string))

		if err != nil {
			return nil, err
		}

		return secret, nil
	}

	return nil, fmt.Errorf("no kid in jwt header")

}
