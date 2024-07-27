package actions

import (
	"log"
	pb "server/resources/proto"
)


func(s *Server) handleLoginMessage(message *pb.Message){
	loginMessage := message.GetLoginMessage()
	switch loginMessage.GetStatus(){
	case pb.LoginMessage_REQUEST_TO_LOGIN:
		username := message.GetFromUsername()
		client := s.GetClient(username) 
		encoded = client.EncodeUsingPubK()
	}
}

func (s *Server) handleMessage(message *pb.Message) {
	
	switch msg := message.GetMessage().(type) {
	case *pb.Message_LoginMessage:
		s.handleLoginMessage(msg)
	default:
		log.Printf("Unknown message type received")
	}
}
