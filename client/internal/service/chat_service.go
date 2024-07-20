package service

import (
	"client/internal/model"
)

type ChatService struct {
	client *model.Client
}

func NewChatService(address string) (*ChatService, error) {
	client, err := model.NewClient(address)
	if err != nil {
		return nil, err
	}
	return &ChatService{client: client}, nil
}

func (s *ChatService) SendMessage(message string) error {
	return s.client.SendMessage(message)
}

func (s *ChatService) ReceiveMessage() (string, error) {
	return s.client.ReceiveMessage()
}
