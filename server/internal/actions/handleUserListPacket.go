package actions

import (
	"fmt"
	"net"
	"server/internal/util"
	pb "server/resources/proto"
)

type UserListMessageHandler struct {
	conn     net.Conn
	userList *map[string]net.Conn
}

func NewUserListMessageHandler(conn net.Conn, userList *map[string]net.Conn) *UserListMessageHandler {
	return &UserListMessageHandler{conn: conn, userList: userList}
}

func (h *UserListMessageHandler) handleMessage(message *pb.Message) error {
	var err error
	var reply *pb.UserListPacket
	registerMessage := message.GetUserListMessage()
	if registerMessage == nil || message.GetFromUsername() == "" {
		return fmt.Errorf("unable to parse user list message")
	}

	switch registerMessage.GetStatus() {
	case pb.UserListPacket_REQUEST_USER_LIST:
		fmt.Println("Received request for user list")
		users := make([]string, 0)
		for user := range *h.userList {
			users = append(users, user)
		}
		reply = &pb.UserListPacket{
			Status: pb.UserListPacket_USER_LIST,
			Users:  users,
		}
	default:
		fmt.Println("invalid register message status")
		reply = &pb.UserListPacket{
			Status: pb.UserListPacket_ERROR,
		}
	}

	err = h.sendUserListMessage(reply)
	return err
}

func (h *UserListMessageHandler) sendUserListMessage(reply *pb.UserListPacket) error {
	message := &pb.Message{
		Source: pb.Message_SERVER,
		Packet: &pb.Message_UserListMessage{
			UserListMessage: reply,
		},
	}

	return util.SendMessage(h.conn, message)
}
