package model

import (
	"crypto"
	"crypto/tls"
)
type Client struct{
	Username string
	PublicKey *crypto.PublicKey
	conn *tls.Conn
}

func NewClient(Username string, PublicKey *crypto.PublicKey) (*Client) {
	return &Client{
        Username: Username,
        PublicKey: PublicKey,
    }
}
// EncryptWithPublicKey encrypts a string using the provided public key and returns a Base64 encoded string
func EncryptWithPublicKey(str string, publicKey *crypto.PublicKey) (string, error) {
    // Parse the public key
}

func (c *Client) EncodeUsingPK(str string) (string, error){
	return EncryptWithPublicKey(str,c.PublicKey)
}


