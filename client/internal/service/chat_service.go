package service

import (
	pb "client/resources/proto"
	"errors"
)

type ChatService struct {
	commService *CommunicationService
}

func NewChatService(commService *CommunicationService) *ChatService {
	return &ChatService{commService: commService}
}

func (s *ChatService) SendMessage(message *pb.Message) error {
	return s.commService.SendMessage(message)
}

func (s *ChatService) ReceiveMessage() (*pb.Message, error) {
	chatMsg := <-s.commService.GetChatChannel()
	return &pb.Message{Packet: &pb.Message_ChatMessage{ChatMessage: chatMsg}}, nil
}

func (s *ChatService) GetUserList() ([]string, error) {
	userListRequest := &pb.Message{
		Source:       pb.Message_CLIENT,
		FromUsername: &s.commService.GetClient().Username,
		Packet: &pb.Message_UserListMessage{
			UserListMessage: &pb.UserListPacket{
				Status: pb.UserListPacket_REQUEST_USER_LIST,
			},
		},
	}

	err := s.commService.SendMessage(userListRequest)
	if err != nil {
		return nil, err
	}

	userListChan := s.commService.GetUserListChannel()
	userListMessage := <-userListChan

	if userListMessage == nil {
		return nil, errors.New("invalid user list response")
	}

	return userListMessage.GetUsers(), nil
}
