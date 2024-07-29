package main

import (
	"crypto/tls"
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"net"
	"server/internal/actions"
	"server/internal/util"
	pb "server/resources/proto"
	"sync"
)

type Server struct {
	address       string
	clients       map[net.Conn]bool
	handlers      map[net.Conn]map[string]actions.MessageHandler
	handlersMutex sync.Mutex
	//
	listOfLoggedInUsers map[string]net.Conn
	loggedInUsersMutex  sync.RWMutex
}

func NewServer(address string) *Server {
	return &Server{
		address:             address,
		clients:             make(map[net.Conn]bool),
		handlers:            make(map[net.Conn]map[string]actions.MessageHandler),
		listOfLoggedInUsers: make(map[string]net.Conn),
	}
}

func (s *Server) getOrCreateHandler(conn net.Conn, handlerType string) actions.MessageHandler {
	s.handlersMutex.Lock()
	defer s.handlersMutex.Unlock()

	if _, exists := s.handlers[conn]; !exists {
		s.handlers[conn] = make(map[string]actions.MessageHandler)
	}

	if handler, exists := s.handlers[conn][handlerType]; exists {
		return handler
	}

	var newHandler actions.MessageHandler
	switch handlerType {
	case "login":
		newHandler = actions.NewLoginMessageHandler(conn, &s.listOfLoggedInUsers)
	case "register":
		newHandler = actions.NewRegisterMessageHandler(conn)
	case "user_list":
		newHandler = actions.NewUserListMessageHandler(conn, &s.listOfLoggedInUsers)
	default:
		log.Printf("Unknown handler type: %s\n", handlerType)
		return nil
	}

	s.handlers[conn][handlerType] = newHandler
	return newHandler
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

func (s *Server) addLoggedInUsers(username string, conn net.Conn) {
	s.loggedInUsersMutex.Lock()
	defer s.loggedInUsersMutex.Unlock()
	s.listOfLoggedInUsers[username] = conn
}

func (s *Server) removeHandlers(conn net.Conn) {
	s.handlersMutex.Lock()
	s.loggedInUsersMutex.Lock()
	defer s.loggedInUsersMutex.Unlock()
	defer s.handlersMutex.Unlock()
	delete(s.handlers, conn)
	for username, connection := range s.listOfLoggedInUsers {
		if connection == conn {
			delete(s.listOfLoggedInUsers, username)
		}
	}
}

func (s *Server) handleClient(conn net.Conn) {
	defer conn.Close()
	defer delete(s.clients, conn)
	defer s.removeHandlers(conn)

	for {
		message, err := util.ReadMessage(conn)
		if err != nil {
			log.Printf("Error reading from client: %v\n", err)
			return
		}

		var messageHandler actions.MessageHandler
		switch message.GetPacket().(type) {
		case *pb.Message_LoginMessage:
			messageHandler = s.getOrCreateHandler(conn, "login")
		case *pb.Message_RegisterMessage:
			messageHandler = s.getOrCreateHandler(conn, "register")
		case *pb.Message_UserListMessage:
			messageHandler = s.getOrCreateHandler(conn, "user_list")
		default:
			log.Printf("Unknown message type: %v\n", message)
			continue
		}

		messageContext := actions.NewMessageContext(messageHandler)
		if err := messageContext.ExecuteStrategy(message); err != nil {
			log.Printf("Error executing strategy: %v\n", err)
			errorMsg := &pb.Message{
				Source: pb.Message_SERVER,
				Packet: &pb.Message_LoginMessage{
					LoginMessage: &pb.LoginPacket{
						Status: pb.LoginPacket_LOGIN_FAILED,
					},
				},
			}
			if sendErr := util.SendMessage(conn, errorMsg); sendErr != nil {
				log.Printf("Error sending error message to client: %v\n", sendErr)
			}
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
