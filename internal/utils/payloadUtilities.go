package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

// GenerateMessageDigestOrSignature | Returns the hex encoded HMAC signature of the message using the key
func GenerateMessageDigestOrSignature(msg, key []byte) (string, error) {

	mac := hmac.New(sha256.New, key)
	_, err := mac.Write(msg)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(mac.Sum(nil)), nil
}

// VerifyPayloadWithHash | Verifies the message with the key and hash
func VerifyPayloadWithHash(msg, key []byte, hash string) (bool, error) {

	sig, err := hex.DecodeString(hash)
	if err != nil {
		return false, err
	}

	mac := hmac.New(sha256.New, key)
	mac.Write(msg)

	return hmac.Equal(sig, mac.Sum(nil)), nil
}
