package actions

import (
	"fmt"
	"google.golang.org/protobuf/proto"
	"net"
	"server/internal/db"
	pb "server/resources/proto"
)

type RegisterMessageHandler struct {
	conn *net.TCPConn
}

func NewRegisterMessageHandler(conn *net.TCPConn) *RegisterMessageHandler {
	return &RegisterMessageHandler{conn: conn}
}

func (h *RegisterMessageHandler) handleMessage(message *pb.Message) error {
	var err error
	registerMessage := message.GetRegisterMessage()
	if registerMessage == nil {
		return fmt.Errorf("unable to parse register message")
	}

	switch registerMessage.GetStatus() {
	case pb.RegisterPacket_REQUEST_TO_REGISTER:
		fmt.Println("Received request to register")
		database := db.GetDatabase()
		if err := database.CreateNewUser(message.GetFromUsername(), registerMessage.GetPublicKey()); err != nil {
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

	data, err := proto.Marshal(message)
	if err != nil {
		return fmt.Errorf("error marshalling message: %v", err)
	}

	_, err = h.conn.Write(data)
	if err != nil {
		return fmt.Errorf("error sending message: %v", err)
	}

	return nil
}
