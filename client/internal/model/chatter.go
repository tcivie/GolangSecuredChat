package model

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
)

type Chatter struct {
	Username  string
	publicKey *rsa.PublicKey
	aes256Key []byte
	//
	cypherBlock *cipher.Block
}

func NewChatter(username string) *Chatter {
	return &Chatter{
		Username: username,
	}
}

func (c *Chatter) SetAES256Key(aes256Key []byte) {
	c.aes256Key = aes256Key
	block, _ := aes.NewCipher(aes256Key)
	c.cypherBlock = &block
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
	encryptedMessage := make([]byte, len(message))
	(*c.cypherBlock).Encrypt(encryptedMessage, []byte(message))
	return encryptedMessage
}

func (c *Chatter) Decrypt(encryptedMessage []byte) string {
	decryptedMessage := make([]byte, len(encryptedMessage))
	(*c.cypherBlock).Decrypt(decryptedMessage, encryptedMessage)
	return string(decryptedMessage)
}
