package model

import (
	pb "client/resources/proto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/binary"
	"fmt"
	"golang.org/x/crypto/ssh"
	"google.golang.org/protobuf/proto"
	"io"
	"log"
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
	key, err := ssh.ParseRawPrivateKey(privateKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("error parsing private key: %v", err)
	}
	return &Client{
		Conn:        conn,
		isConnected: isConnected,
		privateKey:  key.(*rsa.PrivateKey),
	}, err
}

func (c *Client) SendMessage(message *pb.Message) error {
	data, err := proto.Marshal(message)
	if err != nil {
		return err
	}

	// Write the length of the message
	log.Println("SendMessage data length: ", len(data))
	err = binary.Write(c.Conn, binary.BigEndian, uint32(len(data)))
	if err != nil {
		return err
	}

	// Write the message itself
	log.Println("SendMessage data: ", data)
	_, err = c.Conn.Write(data)
	return err
}

func (c *Client) GetMessage() (*pb.Message, error) {
	// Read the message length
	var length uint32
	err := binary.Read(c.Conn, binary.BigEndian, &length)
	log.Println("GetMessage length: ", length)
	if err != nil {
		log.Println("Error reading message length: ", err)
		return nil, err
	}

	// Read the message data
	data := make([]byte, length)
	_, err = io.ReadFull(c.Conn, data)
	log.Println("GetMessage data: ", data)
	if err != nil {
		log.Println("Error reading message data: ", err)
		return nil, err
	}

	// Unmarshal the message
	message := &pb.Message{}
	if err := proto.Unmarshal(data, message); err != nil {
		log.Println("Error unmarshalling message: ", err)
		return nil, err
	}

	return message, nil
}

func (c *Client) Close() error {
	return c.Conn.Close()
}

func (c *Client) DecryptMessageWithPrivateKey(message []byte) ([]byte, error) {
	decrypted, err := rsa.DecryptOAEP(sha256.New(), nil, c.privateKey, message, nil)
	if err != nil {
		return nil, fmt.Errorf("decryption error: %v", err)
	}
	return decrypted, nil
}

func (c *Client) GetPubKey() *rsa.PublicKey {
	pubKey := c.privateKey.Public()
	pubKeyRSA, ok := pubKey.(*rsa.PublicKey)
	if !ok {
		fmt.Println("Error casting public key")
		return nil
	}
	return pubKeyRSA
}
