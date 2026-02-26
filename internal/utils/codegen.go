package utils

import (
	"crypto/rand"
	"math/big"
)

const urlCodeCharset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const urlCodeLength = 11

// GenerateURLCode generates a random 11-character alphanumeric code
// using uppercase letters and digits (e.g., "WD969DLF05F").
func GenerateURLCode() string {
	code := make([]byte, urlCodeLength)
	max := big.NewInt(int64(len(urlCodeCharset)))
	for i := range code {
		n, _ := rand.Int(rand.Reader, max)
		code[i] = urlCodeCharset[n.Int64()]
	}
	return string(code)
}

// GenerateVerificationToken generates a random 64-character hex token
// for email verification.
func GenerateVerificationToken() string {
	const charset = "abcdef0123456789"
	token := make([]byte, 64)
	max := big.NewInt(int64(len(charset)))
	for i := range token {
		n, _ := rand.Int(rand.Reader, max)
		token[i] = charset[n.Int64()]
	}
	return string(token)
}
