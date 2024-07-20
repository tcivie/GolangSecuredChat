package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

type Client struct {
	conn net.Conn
}

func NewClient(address string) (*Client, error) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, fmt.Errorf("error connecting to server: %v", err)
	}

	return &Client{conn: conn}, nil
}

func (c *Client) Start() {
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
	client, err := NewClient("localhost:8080")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Connected to chat server. Start typing messages:")
	client.Start()
}
