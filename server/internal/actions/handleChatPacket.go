package actions

import (
	"fmt"
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
	if message == nil {
		return fmt.Errorf("received nil message")
	}

	// Check if the message has the ChatMessage field
	if message.Packet == nil {
		return fmt.Errorf("message packet is nil")
	}

	chatMessage, ok := message.Packet.(*pb.Message_ChatMessage)
	if !ok {
		return fmt.Errorf("message is not a ChatMessage")
	}

	if chatMessage.ChatMessage == nil {
		return fmt.Errorf("ChatMessage is nil")
	}

	toUsername := chatMessage.ChatMessage.GetToUsername()
	if toUsername == "" {
		return fmt.Errorf("recipient username is empty")
	}

	toConn, exists := (*cmh.listOfLoggedInUsers)[toUsername]
	if !exists || toConn == nil {
		return fmt.Errorf("recipient %s is not logged in or connection is nil", toUsername)
	}

	// Forward the message as is to the recipient
	if err := util.SendMessage(toConn, message); err != nil {
		return fmt.Errorf("failed to send message to %s: %v", toUsername, err)
	}

	return nil
}
