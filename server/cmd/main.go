package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"server/internal/model"
	pb "server/resources/proto"

	"google.golang.org/protobuf/proto"
)

type Server struct {
	address string
	clients map[net.Conn]bool
}

func (s *Server) GetClient(username string) *model.Client{
	return nil
}

func NewServer(address string) *Server {
	return &Server{
		address: address,
		clients: make(map[net.Conn]bool),
	}
}

func (s *Server) Start() error {
	// Load the server certificate and private key
	cert, err := tls.LoadX509KeyPair("resources/auth/server-cert.pem", "resources/auth/server-key.pem")
	if err != nil {
		return fmt.Errorf("error loading server certificate: %v", err)
	}

	// Create the TLS configuration
	config := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12, // Ensure minimum TLS version 1.2
	}

	// Create a TLS listener
	listener, err := tls.Listen("tcp", s.address, config)
	if err != nil {
		return fmt.Errorf("error starting TLS server: %v", err)
	}

	defer func(listener net.Listener) {
		err := listener.Close()
		if err != nil {
			log.Printf("Error closing listener: %v\n", err)
		}
	}(listener)

	log.Printf("TLS Server listening on %s\n", s.address)
	s.handleConnections(listener)

	return nil
}

func (s *Server) handleConnections(listener net.Listener) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %v\n", err)
			continue
		}

		s.clients[conn] = true
		go s.handleClient(conn)
	}
}

func (s *Server) handleClient(conn net.Conn) {
	defer conn.Close()
	defer delete(s.clients, conn)

	for {
		buffer := make([]byte, 1024*4)
		n, err := conn.Read(buffer)
		if err != nil {
			log.Printf("Error reading from client: %v\n", err)
			return
		}

		message := &pb.Message{}
		if err := proto.Unmarshal(buffer[:n], message); err != nil {
			log.Fatalln("Failed to parse Message:", err)
		}

		//s.broadcast(message, conn)
	}
}

func (s *Server) broadcast(message []byte, sender net.Conn) {
	for client := range s.clients {
		if client == sender {
			continue
		}

		_, err := client.Write(message)
		if err != nil {
			log.Printf("Error broadcasting to client: %v\n", err)
			client.Close()
			delete(s.clients, client)
		}
	}
}

func main() {
	server := NewServer(":8080")
	log.Fatal(server.Start())
}
