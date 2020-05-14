package bcrypt

import (
	"bytes"
	"crypto/rand"
	"crypto/sha1"
	"encoding/hex"
	"io"
)

// SaltSize is salt size in bytes.
const SaltSize = 16

// EncodePassword ...
func EncodePassword(password string) string {
	return hexa(saltedHashPassword([]byte(password)))
}

// CompareHashAndPassword
func CompareHashAndPassword(hashpwd, password string) bool {
	return isPasswordMatch(dehexa(hashpwd), []byte(password))
}

func saltedHashPassword(secret []byte) []byte {
	buf := make([]byte, SaltSize, SaltSize+sha1.Size)
	_, err := io.ReadFull(rand.Reader, buf)
	if err != nil {
		panic(err)
	}

	h := sha1.New()
	_, err = h.Write(buf)
	if err != nil {

	}

	_, err = h.Write(secret)
	if err != nil {
	}

	return h.Sum(buf)
}

func isPasswordMatch(data, secret []byte) bool {
	if len(data) != SaltSize+sha1.Size {
		panic("wrong length of data")
	}

	h := sha1.New()
	_, err := h.Write(data[:SaltSize])
	if err != nil {
	}

	_, err = h.Write(secret)
	if err != nil {
	}

	return bytes.Equal(h.Sum(nil), data[SaltSize:])
}

func hexa(data []byte) string {
	return hex.EncodeToString(data)
}

func dehexa(s string) []byte {
	b, _ := hex.DecodeString(s)
	return b
}
