package actions

import (
	"net"
	"server/internal/util"
	pb "server/resources/proto"
)

type ChatMessageHandler struct {
	listOfLoggedInUsers *map[string]net.Conn
}

func NewChatMessageHandler(listOfLoggedInUsers *map[string]net.Conn) *ChatMessageHandler {
	return &ChatMessageHandler{listOfLoggedInUsers: listOfLoggedInUsers}
}

func (cmh *ChatMessageHandler) handleMessage(message *pb.Message) error {
	chatMessage := message.GetChatMessage()
	if chatMessage == nil {
		return nil
	}
	// Forward the message as is to the recipient
	toConn := (*cmh.listOfLoggedInUsers)[chatMessage.GetToUsername()]
	if toConn == nil {
		return nil //TODO: Handle this case
	}
	return util.SendMessage(toConn, message)
}
