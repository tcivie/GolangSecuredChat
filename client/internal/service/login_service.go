package service

import (
	"client/internal/model"
	pb "client/resources/proto"
)

type LoginService struct {
	client *model.Client
}

func NewLoginService(client *model.Client) *LoginService {
	return &LoginService{
		client: client,
	}
}

func (ls *LoginService) Login() error {
	loginState := &pb.LoginPacket{
		Status: pb.LoginPacket_REQUEST_TO_LOGIN,
	}
	message := &pb.Message{
		Source:       pb.Message_CLIENT,
		FromUsername: &ls.client.Username,
		Packet:       &pb.Message_LoginMessage{LoginMessage: loginState},
	}
	return ls.client.SendMessage(message)
}
