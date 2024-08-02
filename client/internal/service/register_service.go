package service

import (
	pb "client/resources/proto"
	"errors"
)

type RegisterService struct {
	commService *CommunicationService
}

func NewRegisterService(commService *CommunicationService) *RegisterService {
	return &RegisterService{
		commService: commService,
	}
}

func (rs *RegisterService) Register(username string) error {
	pubKey := rs.commService.GetClient().GetPubKey()
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

	if err := rs.commService.SendMessage(message); err != nil {
		return err
	}

	// Wait for response on the register channel
	registerChan := rs.commService.GetRegisterChannel()
	registerMessage := <-registerChan

	if registerMessage == nil || registerMessage.GetStatus() != pb.RegisterPacket_REGISTER_SUCCESS {
		return errors.New("invalid register message")
	}
	return nil
}
