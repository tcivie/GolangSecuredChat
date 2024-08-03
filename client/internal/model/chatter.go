package model

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"errors"
)

type Chatter struct {
	Username    string
	publicKey   *rsa.PublicKey
	aes256Key   []byte
	cypherBlock cipher.Block
}

func NewChatter(username string) *Chatter {
	return &Chatter{
		Username: username,
	}
}

func (c *Chatter) SetAES256Key(aes256Key []byte) {
	c.aes256Key = aes256Key
	block, _ := aes.NewCipher(aes256Key)
	c.cypherBlock = block
}

func (c *Chatter) SetPublicKey(publicKey *rsa.PublicKey) {
	c.publicKey = publicKey
}

func (c *Chatter) IsHandShaken() bool {
	return c.publicKey != nil
}

func (c *Chatter) EncryptWithPublicKey(message []byte) ([]byte, error) {
	return rsa.EncryptOAEP(sha256.New(), rand.Reader, c.publicKey, message, nil)
}

func (c *Chatter) GetPubKey() *rsa.PublicKey {
	return c.publicKey
}

func (c *Chatter) Encrypt(message string) []byte {
	plaintext := []byte(message)
	plaintext = pkcs7Pad(plaintext, aes.BlockSize)
	ciphertext := make([]byte, len(plaintext))
	for i := 0; i < len(plaintext); i += aes.BlockSize {
		c.cypherBlock.Encrypt(ciphertext[i:i+aes.BlockSize], plaintext[i:i+aes.BlockSize])
	}
	return ciphertext
}

func (c *Chatter) Decrypt(encryptedMessage []byte) string {
	plaintext := make([]byte, len(encryptedMessage))
	for i := 0; i < len(encryptedMessage); i += aes.BlockSize {
		c.cypherBlock.Decrypt(plaintext[i:i+aes.BlockSize], encryptedMessage[i:i+aes.BlockSize])
	}
	unpaddedPlaintext, err := pkcs7Unpad(plaintext, aes.BlockSize)
	if err != nil {
		return "" // Handle error appropriately in your application
	}
	return string(unpaddedPlaintext)
}

// pkcs7Pad adds PKCS#7 padding to the data
func pkcs7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - (len(data) % blockSize)
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padText...)
}

// pkcs7Unpad removes PKCS#7 padding from the data
func pkcs7Unpad(data []byte, blockSize int) ([]byte, error) {
	length := len(data)
	if length == 0 {
		return nil, errors.New("invalid padding")
	}
	if length%blockSize != 0 {
		return nil, errors.New("invalid padding")
	}
	padding := int(data[length-1])
	if padding > blockSize || padding == 0 {
		return nil, errors.New("invalid padding")
	}
	for i := length - padding; i < length; i++ {
		if int(data[i]) != padding {
			return nil, errors.New("invalid padding")
		}
	}
	return data[:length-padding], nil
}
