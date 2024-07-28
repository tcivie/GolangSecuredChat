package service

import (
	"client/internal/model"
	pb "client/resources/proto"
	"errors"
)

type RegisterService struct {
	client *model.Client
}

func NewRegisterService(client *model.Client) *RegisterService {
	return &RegisterService{
		client: client,
	}
}

func (rs *RegisterService) Register(username string) error {
	pubKey := rs.client.GetPubKey()
	if pubKey == nil {
		return errors.New("public key not found")
	}
	// Create a register packet
	registerState := &pb.RegisterPacket{
		Status:    pb.RegisterPacket_REQUEST_TO_REGISTER,
		PublicKey: pubKey.N.Bytes(),
	}
	message := &pb.Message{
		Source:       pb.Message_CLIENT,
		FromUsername: &username,
		Packet:       &pb.Message_RegisterMessage{RegisterMessage: registerState},
	}

	if err := rs.client.SendMessage(message); err != nil {
		return err
	}

	message, err := rs.client.GetMessage()
	if err != nil {
		return err
	}
	registerMessage := message.GetRegisterMessage()
	if registerMessage == nil || registerMessage.GetStatus() != pb.RegisterPacket_REGISTER_SUCCESS {
		return errors.New("invalid register message")
	}
	return nil
}
