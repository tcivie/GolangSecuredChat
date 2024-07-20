package model

import (
	"bufio"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
)

type Client struct {
	conn *tls.Conn
}

func NewClient(address string) (*Client, error) {
	cert, err := os.ReadFile("resources/server-cert.pem")
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
	if err != nil {
		return nil, fmt.Errorf("error connecting to server: %v", err)
	}

	return &Client{conn: conn}, nil
}

func (c *Client) SendMessage(message string) error {
	_, err := fmt.Fprintf(c.conn, message+"\n")
	return err
}

func (c *Client) ReceiveMessage() (string, error) {
	return bufio.NewReader(c.conn).ReadString('\n')
}

func (c *Client) Close() error {
	return c.conn.Close()
}
