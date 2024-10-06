package toolcrypto

import (
	"crypto/rand"
	"encoding/base32"
	"encoding/base64"

	"golang.org/x/crypto/bcrypt"
)

func RandomCryptoArray(n int) ([]byte, error) {
	// Create a byte slice to hold the random bytes
	byteSlice := make([]byte, n)

	// Fill the byte slice with random bytes
	_, err := rand.Read(byteSlice)
	if err != nil {
		// panic("cannot create random crypto string")
		return nil, err
	}

	return byteSlice, nil
}
func RandomCryptoBase64(n int) (string, error) {

	byteSlice, err := RandomCryptoArray(n)
	if err != nil {
		return "", err
	}
	// Encode the byte slice to a base64 string and return it
	return base64.StdEncoding.EncodeToString(byteSlice), nil // , nil
}
func RandomCryptoBase32(n int) (string, error) {

	byteSlice, err := RandomCryptoArray(n)
	if err != nil {
		return "", err
	}
	// Encode the byte slice to a base64 string and return it
	return base32.StdEncoding.EncodeToString(byteSlice), nil // , nil
}

// HashPassword hashes the password using bcrypt
func HashPassword(password string) (string, error) {
	// Generate a bcrypt hash of the password with a default cost
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	// Return the hashed password as a string
	return string(hash), nil
}

// CompareHashAndPassword compares the hashed password with the plain text password
func CompareHashAndPassword(hash, password string) bool {

	if hash == "" {
		return false
	}

	if password == "" {
		return false
	}

	// Compare the hashed password with the plain text password
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
