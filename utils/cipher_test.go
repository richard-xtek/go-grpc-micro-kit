package utils

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	message    = "This is a message"
	passphrase = "Top secret"
	secret     = "146526dc365043cf6f6cbe979f2e0d7181be87cf0250c404abf72bd7bd1b0cc56e6cf69fd5ad7ac0f176867912"
)

func TestEncryptDecrypt(t *testing.T) {
	byteMessage := []byte(message)
	encrypt, err := Encrypt(byteMessage, passphrase)
	assert.Nil(t, err)

	decrypt, err := Decrypt(encrypt, passphrase)
	assert.Nil(t, err)
	assert.Equal(t, decrypt, byteMessage)
}

func TestEncryptSuccess(t *testing.T) {
	encrypt, err := Encrypt([]byte(message), passphrase)
	assert.Nil(t, err)

	fmt.Println(hex.EncodeToString(encrypt))
}

func TestDecryptSuccess(t *testing.T) {
	encrypt, err := hex.DecodeString(secret)
	assert.Nil(t, err)

	decrypt, err := Decrypt(encrypt, passphrase)
	assert.Nil(t, err)

	fmt.Println(string(decrypt))
}

func TestDecryptFail(t *testing.T) {
	encrypt, err := hex.DecodeString(secret + "ab")
	assert.Nil(t, err)

	decrypt, err := Decrypt(encrypt, passphrase)
	assert.Equal(t, err.Error(), "cipher: message authentication failed")
	assert.Nil(t, decrypt)
}

func TestDecryptTooShort(t *testing.T) {
	encrypt, err := hex.DecodeString(secret[:22])
	assert.Nil(t, err)

	decrypt, err := Decrypt(encrypt, passphrase)
	assert.Equal(t, err.Error(), "ciphertext too short")
	assert.Nil(t, decrypt)
}
