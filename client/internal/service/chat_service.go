package service

import (
	pb "client/resources/proto"
	"context"
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

func (s *ChatService) ReceiveMessage(ctx context.Context) (*pb.Message, error) {
	select {
	case <-ctx.Done():
		return nil, nil
	case chatMsg := <-s.commService.GetChatChannel():
		return chatMsg, nil
	}
}

func (s *ChatService) GetUserList() ([]string, error) {
	username := s.commService.GetUsername()
	userListRequest := &pb.Message{
		Source:       pb.Message_CLIENT,
		FromUsername: &username,
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
