package service

import (
	"encoding/base64"
	"go-auth/internal/config"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewVaultKey(t *testing.T) {

	decodeBase64 := func(s string) []byte {
		b, _ := base64.StdEncoding.DecodeString(s)
		return b
	}

	vaultKey, err := NewVaultKey()
	assert.NoError(t, err)
	assert.NotNil(t, vaultKey)
	assert.Equal(t, 64, len(decodeBase64(vaultKey.AuthKey)))
	assert.Equal(t, 64, len(decodeBase64(vaultKey.OtpKey)))
	assert.Equal(t, 64, len(decodeBase64(vaultKey.HashKey)))
}

func TestVaultKeyIsEmpty(t *testing.T) {

	vaultKey := VaultKey{}
	assert.True(t, vaultKey.IsEmpty())

	vaultKey.fill()
	assert.False(t, vaultKey.IsEmpty())
}

func TestDefaultVaultService_CurrentKey(t *testing.T) {

	vaultService, err := newVaultService(appService)
	assert.NoError(t, err)

	// Ensure we have at least one key
	key, err := vaultService.CurrentKey()
	assert.NoError(t, err)
	assert.NotNil(t, key)
}

func TestDefaultVaultService_KeyByID(t *testing.T) {

	vaultService, err := newVaultService(appService)
	assert.NoError(t, err)

	// Create a new VaultKey and add it to the vaultService
	vaultKey, _ := NewVaultKey()
	assert.NotNil(t, vaultKey)

	// Manually add to keychain for testing
	(vaultService.(*defaultVaultService)).keychain = append((vaultService.(*defaultVaultService)).keychain, SecretKey{
		ID:      vaultKey.ID,
		AuthKey: []byte("authkey"),
		OtpKey:  []byte("otpkey"),
		HashKey: []byte("hashkey"),
	})

	secretKey, err := vaultService.KeyByID(vaultKey.ID)
	assert.NoError(t, err)
	assert.Equal(t, vaultKey.ID, secretKey.ID)
}

func TestVaultKeyScope_CurrentKey(t *testing.T) {

	vaultService, err := newVaultService(appService)
	assert.NoError(t, err)

	scope := vaultService.KeyScopeAuth()
	keyID, secret, err := scope.CurrentKey()
	assert.NoError(t, err)
	assert.NotEmpty(t, keyID)
	assert.NotNil(t, secret)
}

func TestVaultKeyScope_KeyByID(t *testing.T) {

	vaultService, err := newVaultService(appService)
	assert.NoError(t, err)

	scope := vaultService.KeyScopeOtp()

	vaultKey, _ := NewVaultKey()
	(vaultService.(*defaultVaultService)).keychain = append((vaultService.(*defaultVaultService)).keychain, SecretKey{
		ID:      vaultKey.ID,
		AuthKey: []byte("authkey"),
		OtpKey:  []byte("otpkey"),
		HashKey: []byte("hashkey"),
	})

	secret, err := scope.KeyByID(vaultKey.ID)
	assert.NoError(t, err)
	assert.NotNil(t, secret)
}

func TestAllKeys(t *testing.T) {

	keys, err := allKeys(appService)
	assert.NoError(t, err)
	assert.NotNil(t, keys)

	// Example to check if returned keys are valid.
	for _, key := range keys {
		decodedAuthKey, err := base64.StdEncoding.DecodeString(key.AuthKey)
		assert.NoError(t, err)
		assert.NotNil(t, decodedAuthKey)
	}
}

func TestLoadKeys(t *testing.T) {

	vaultService, err := newVaultService(appService)
	assert.NoError(t, err)

	// Load keys directly for testing
	keys := []config.AppConfigVaultKey{
		{ID: "key1", AuthKey: base64.StdEncoding.EncodeToString([]byte("authkey1")), OtpKey: base64.StdEncoding.EncodeToString([]byte("otpkey1")), HashKey: base64.StdEncoding.EncodeToString([]byte("hashkey1"))},
	}

	(vaultService.(*defaultVaultService)).keychain = nil

	err = (vaultService.(*defaultVaultService)).loadKeys(keys)
	assert.NoError(t, err)
	assert.Equal(t, 1, len((vaultService.(*defaultVaultService)).keychain))
}
