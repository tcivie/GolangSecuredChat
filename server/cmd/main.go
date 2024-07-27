package main

import (
	"crypto/tls"
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"net"
	"server/internal/actions"
	pb "server/resources/proto"

	"google.golang.org/protobuf/proto"
)

type Server struct {
	address string
	clients map[*net.TCPConn]bool
}

func NewServer(address string) *Server {
	return &Server{
		address: address,
		clients: make(map[*net.TCPConn]bool),
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
	tcpListener, ok := listener.(*net.TCPListener)
	if !ok {
		return fmt.Errorf("error converting listener to TCP listener")
	}
	s.handleConnections(tcpListener)

	return nil
}

func (s *Server) handleConnections(listener *net.TCPListener) {
	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			log.Printf("Error accepting connection: %v\n", err)
			continue
		}

		s.clients[conn] = true
		go s.handleClient(conn)
	}
}

func (s *Server) handleClient(conn *net.TCPConn) {
	var messageHandler actions.MessageHandler
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

		switch message.GetPacket().(type) {
		case *pb.Message_LoginMessage:
			messageHandler = actions.NewLoginMessageHandler(conn)
			break
		case *pb.Message_RegisterMessage:
			messageHandler = actions.NewRegisterMessageHandler(conn)
			break
		default:
			log.Printf("Unknown message type: %v\n", message)
		}
		messageContext := actions.NewMessageContext(messageHandler)
		if err := messageContext.ExecuteStrategy(message); err != nil {
			log.Printf("Error executing strategy: %v\n", err)
		}
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
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}
	server := NewServer(":8080")
	log.Fatal(server.Start())
}
