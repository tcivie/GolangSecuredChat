package actions

import (
	"fmt"
	"net"
	"server/internal/db"
	"server/internal/util"
	pb "server/resources/proto"
)

type RegisterMessageHandler struct {
	conn net.Conn
}

func NewRegisterMessageHandler(conn net.Conn) *RegisterMessageHandler {
	return &RegisterMessageHandler{conn: conn}
}

func (h *RegisterMessageHandler) handleMessage(message *pb.Message) error {
	var err error
	registerMessage := message.GetRegisterMessage()
	if registerMessage == nil || message.GetFromUsername() == "" {
		return fmt.Errorf("unable to parse register message")
	}

	switch registerMessage.GetStatus() {
	case pb.RegisterPacket_REQUEST_TO_REGISTER:
		fmt.Println("Received request to register")
		database := db.GetDatabase()
		hashedUsername := util.HashString(message.GetFromUsername())
		if err := database.CreateNewUser(hashedUsername, registerMessage.GetPublicKey()); err != nil {
			registerMessage = &pb.RegisterPacket{
				Status: pb.RegisterPacket_REGISTER_FAILED,
			}
		} else {
			registerMessage = &pb.RegisterPacket{
				Status: pb.RegisterPacket_REGISTER_SUCCESS,
			}
		}
		err = h.sendRegisterMessage(registerMessage)
		break
	default:
		return fmt.Errorf("invalid register message status")
	}
	return err
}

func (h *RegisterMessageHandler) sendRegisterMessage(reply *pb.RegisterPacket) error {
	message := &pb.Message{
		Source: pb.Message_SERVER,
		Packet: &pb.Message_RegisterMessage{
			RegisterMessage: reply,
		},
	}

	return util.SendMessage(h.conn, message)
}
