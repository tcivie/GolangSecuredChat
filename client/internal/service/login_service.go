package service

import (
	"client/internal/model"
	pb "client/resources/proto"
	"client/util"
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

func (ls *LoginService) Login() error {
	loginState := &pb.LoginPacket{
		Status: pb.LoginPacket_REQUEST_TO_LOGIN,
	}
	message := &pb.Message{
		Source:       pb.Message_CLIENT,
		FromUsername: &ls.client.Username,
		Packet:       &pb.Message_LoginMessage{LoginMessage: loginState},
	}
	if err := ls.client.SendMessage(message); err != nil {
		return err
	}

	message, err := util.GetMessage(ls.client.Conn, message)
	if err != nil {
		return err
	}

	loginMessage := message.GetLoginMessage() // Encrypted Token
	if loginMessage == nil || loginMessage.GetStatus() != pb.LoginPacket_ENCRYPTED_TOKEN {
		return errors.New("invalid login message")
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
		FromUsername: &ls.client.Username,
		Packet:       &pb.Message_LoginMessage{LoginMessage: loginState},
	}
	if err := ls.client.SendMessage(message); err != nil {
		return err
	}

	message, err = util.GetMessage(ls.client.Conn, message)
	if err != nil {
		return err
	}

	loginMessage = message.GetLoginMessage()
	if loginMessage == nil || loginMessage.GetStatus() != pb.LoginPacket_LOGIN_SUCCESS {
		return errors.New("invalid login message")
	}
	return nil
}
