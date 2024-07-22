package model

import (
	"crypto/rsa"
	"crypto/tls"
)

type Client struct {
	Username  string
	publicKey *rsa.PublicKey
	conn      *tls.Conn
}

func NewClient(username string, publicKey *rsa.PublicKey) *Client {
	return &Client{
		Username:  username,
		publicKey: publicKey,
	}
}

func (c *Client) EncodeUsingPubK(msg string) ([]byte, error) {
	return rsa.EncryptPKCS1v15(nil, c.publicKey, []byte(msg))
}
