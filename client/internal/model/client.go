package model

import (
	pb "client/resources/proto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"google.golang.org/protobuf/proto"
	"os"
)

type Client struct {
	Conn        *tls.Conn
	Username    string
	isLoggedIn  bool
	isConnected bool
	//
	privateKey *rsa.PrivateKey
}

func NewClient(address string, privateKeyPath string) (*Client, error) {
	cert, err := os.ReadFile("resources/auth/server-cert.pem")
	if err != nil {
		return nil, fmt.Errorf("error reading server certificate: %v", err)
	}

	certPool := x509.NewCertPool()
	if ok := certPool.AppendCertsFromPEM(cert); !ok {
		return nil, fmt.Errorf("failed to append certificate")
	}

	config := &tls.Config{
		RootCAs: certPool,
	}

	conn, err := tls.Dial("tcp", address, config)
	isConnected := true
	if err != nil {
		err = fmt.Errorf("error connecting to server: %v", err)
		isConnected = false
	}
	//
	privateKeyBytes, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("error reading private key: %v", err)
	}
	privateKey, err := x509.ParsePKCS1PrivateKey(privateKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("error parsing private key: %v", err)
	}

	return &Client{
		Conn:        conn,
		isConnected: isConnected,
		privateKey:  privateKey,
	}, err
}

func (c *Client) SendMessage(message *pb.Message) error {
	data, err := proto.Marshal(message)
	if err != nil {
		return err
	}
	_, err = c.Conn.Write(data)
	return err
}

func (c *Client) ReceiveMessage() (*pb.Message, error) {
	data := make([]byte, 1024*4)
	n, err := c.Conn.Read(data)
	if err != nil {
		return nil, err
	}
	message := pb.Message{}
	if err := proto.Unmarshal(data[:n], &message); err != nil {
		return nil, err
	}
	return &message, nil
}

func (c *Client) Close() error {
	return c.Conn.Close()
}

func (c *Client) DecryptMessage(message []byte) ([]byte, error) {
	decrypted, err := rsa.DecryptOAEP(sha256.New(), nil, c.privateKey, message, nil)
	if err != nil {
		return nil, fmt.Errorf("decryption error: %v", err)
	}
	return decrypted, nil
}
