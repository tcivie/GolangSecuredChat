package service

import (
	"client/internal/model"
	pb "client/resources/proto"
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
