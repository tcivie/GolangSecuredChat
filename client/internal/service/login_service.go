package service

import (
	"client/internal/model"
	pb "client/resources/proto"
	"errors"
)

type LoginService struct {
	client *model.Client
}

func NewLoginService(client *model.Client) *LoginService {
	return &LoginService{
		client: client,
	}
}

func (ls *LoginService) Login(username string) error {
	loginState := &pb.LoginPacket{
		Status: pb.LoginPacket_REQUEST_TO_LOGIN,
	}
	message := &pb.Message{
		Source:       pb.Message_CLIENT,
		FromUsername: &username,
		Packet:       &pb.Message_LoginMessage{LoginMessage: loginState},
	}
	if err := ls.client.SendMessage(message); err != nil {
		return err
	}

	message, err := ls.client.GetMessage()
	if err != nil {
		return err
	}

	loginMessage := message.GetLoginMessage() // Encrypted Token
	if loginMessage == nil || loginMessage.GetStatus() != pb.LoginPacket_ENCRYPTED_TOKEN {
		return errors.New("invalid login")
	}

	// Decrypt token
	decryptedToken, err := ls.client.DecryptMessage(loginMessage.GetToken())
	if err != nil {
		return err
	}

	// Send decrypted token to server
	loginState = &pb.LoginPacket{
		Status: pb.LoginPacket_DECRYPTED_TOKEN,
		Token:  decryptedToken,
	}
	message = &pb.Message{
		Source:       pb.Message_CLIENT,
		FromUsername: &username,
		Packet:       &pb.Message_LoginMessage{LoginMessage: loginState},
	}
	if err := ls.client.SendMessage(message); err != nil {
		return err
	}

	message, err = ls.client.GetMessage()
	if err != nil {
		return err
	}

	loginMessage = message.GetLoginMessage()
	if loginMessage == nil || loginMessage.GetStatus() != pb.LoginPacket_LOGIN_SUCCESS {
		return errors.New("invalid login")
	}
	return nil
}
