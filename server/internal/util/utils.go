package util

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
)

func HashString(data string) string {
	combined := os.Getenv("SERVER_HASH_PASSWORD") + data + os.Getenv("SERVER_HASH_SALT")
	hasher := sha256.New()

	hasher.Write([]byte(combined))
	hashedBytes := hasher.Sum(nil)
	return base64.StdEncoding.EncodeToString(hashedBytes)
}

func EncodeUsingPubK(msgBytes []byte, pubKey *rsa.PublicKey) ([]byte, error) {
	if pubKey == nil {
		return nil, errors.New("public key is nil")
	}
	maxMsgLen := pubKey.Size() - 2*sha256.Size - 2
	if len(msgBytes) > maxMsgLen {
		return nil, fmt.Errorf("message too long: max length is %d bytes", maxMsgLen)
	}
	encrypted, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, pubKey, msgBytes, nil)
	if err != nil {
		return nil, fmt.Errorf("encryption error: %v", err)
	}
	return encrypted, nil
}

func GenerateRandomToken(length int) ([]byte, error) {
	randomBytes := make([]byte, length)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return nil, fmt.Errorf("error generating random bytes: %v", err)
	}

	return randomBytes, nil
}

func DebugPrintPublicKey(key *rsa.PublicKey) string {
	return fmt.Sprintf("N: %x... (len: %d bits), E: %d", key.N.Bytes()[:20], key.N.BitLen(), key.E)
}
