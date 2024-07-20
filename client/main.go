package main

import (
	"bufio"
	"crypto"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"os"
)

type Client struct {
	conn        *tls.Conn
	username    string
	privateKey  crypto.PrivateKey
	publicKey   crypto.PublicKey
	isConnected bool
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

func (c *Client) Start() {
	defer func(conn *tls.Conn) {
		err := conn.Close()
		if err != nil {
			log.Printf("Error closing connection: %v\n", err)
		}
	}(c.conn)

	go c.receiveMessages()
	c.sendMessages()

}

func (c *Client) receiveMessages() {
	for {
		message, err := bufio.NewReader(c.conn).ReadString('\n')
		if err != nil {
			log.Printf("Error receiving message: %v\n", err)
			return
		}
		fmt.Print("Received: " + message)
	}
}

func (c *Client) sendMessages() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		message := scanner.Text()
		_, err := fmt.Fprintf(c.conn, message+"\n")
		if err != nil {
			log.Printf("Error sending message: %v\n", err)
			return
		}
	}
}

func main() {
	client, err := NewClient("localhost:8080") // TODO: store in a file or get from user input
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Connected to chat server. Start typing messages:")
	client.Start()
}
