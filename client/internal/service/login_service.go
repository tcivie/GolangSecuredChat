package service

import (
	pb "client/resources/proto"
	"errors"
)

type LoginService struct {
	commService *CommunicationService
	username    string
}

func NewLoginService(commService *CommunicationService) *LoginService {
	return &LoginService{
		commService: commService,
	}
}

func (ls *LoginService) Login(username string) error {
	ls.username = username
	loginState := &pb.LoginPacket{
		Status: pb.LoginPacket_REQUEST_TO_LOGIN,
	}
	message := &pb.Message{
		Source:       pb.Message_CLIENT,
		FromUsername: &username,
		Packet:       &pb.Message_LoginMessage{LoginMessage: loginState},
	}
	if err := ls.commService.SendMessage(message); err != nil {
		return err
	}

	// Wait for response on the login channel
	loginChan := ls.commService.GetLoginChannel()
	loginMessage := <-loginChan

	if loginMessage == nil || loginMessage.GetStatus() != pb.LoginPacket_ENCRYPTED_TOKEN {
		return errors.New("invalid login")
	}

	// Decrypt token
	decryptedToken, err := ls.commService.GetClient().DecryptMessageWithPrivateKey(loginMessage.GetToken())
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
	if err := ls.commService.SendMessage(message); err != nil {
		return err
	}

	// Wait for final login response
	loginMessage = <-loginChan

	if loginMessage == nil || loginMessage.GetStatus() != pb.LoginPacket_LOGIN_SUCCESS {
		return errors.New("invalid login")
	}
	ls.commService.SetClientUsername(username)
	return nil
}
