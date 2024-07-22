package actions

import (
	"log"
	pb "server/resources/proto"
)

type Server struct {
	address string
}

func (s *Server) handleMessage(message *pb.Message) {
	switch msg := message.GetMessage().(type) {
	case *pb.Message_LoginMessage:
		msg.LoginMessage.GetToken()
	default:
		log.Printf("Unknown message type received")
	}
}
