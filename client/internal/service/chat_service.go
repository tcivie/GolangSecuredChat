package service

import (
	"client/internal/model"
	pb "client/resources/proto"
	"errors"
)

type ChatService struct {
	Client *model.Client
}

func NewChatService(address string, privateKeyPath string) (*ChatService, error) {
	client, err := model.NewClient(address, privateKeyPath)
	if err != nil {
		return nil, err
	}
	return &ChatService{Client: client}, nil
}

func (s *ChatService) SendMessage(message *pb.Message) error {
	return s.Client.SendMessage(message)
}

func (s *ChatService) ReceiveMessage() (*pb.Message, error) {
	return s.Client.GetMessage()
}

func (s *ChatService) GetUserList() ([]string, error) {
	userListRequest := &pb.Message{
		Source:       pb.Message_CLIENT,
		FromUsername: &s.Client.Username,
		Packet: &pb.Message_UserListMessage{
			UserListMessage: &pb.UserListPacket{
				Status: pb.UserListPacket_REQUEST_USER_LIST,
			},
		},
	}

	err := s.Client.SendMessage(userListRequest)
	if err != nil {
		return nil, err
	}

	response, err := s.Client.GetMessage()
	if err != nil {
		return nil, err
	}

	userListMessage := response.GetUserListMessage()
	if userListMessage == nil {
		return nil, errors.New("invalid user list response")
	}

	return userListMessage.GetUsers(), nil
}
