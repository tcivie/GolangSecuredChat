package viewmodel

import (
	"client/internal/model"
	"client/internal/service"
	pb "client/resources/proto"
	"fmt"
)

type ChatViewModel struct {
	chatService *service.ChatService
	messages    []model.Message
}

func NewChatViewModel(service *service.ChatService) *ChatViewModel {
	return &ChatViewModel{
		chatService: service,
		messages:    []model.Message{},
	}
}

func (vm *ChatViewModel) SendMessage(content string) {
	chatMessage := &pb.Message{
		Source:       pb.Message_CLIENT,
		FromUsername: &vm.chatService.Client.Username,
		Packet: &pb.Message_ChatMessage{
			ChatMessage: &pb.ChatPacket{
				ToUsername: "test", // TODO: Implement this
				Message:    content,
			},
		},
	}
	err := vm.chatService.SendMessage(chatMessage)
	if err != nil {
		vm.AddMessage(model.Message{Content: "Error sending message: " + err.Error(), Sender: "System"})
	} else {
		vm.AddMessage(model.Message{Content: content, Sender: "You"})
	}
}

func (vm *ChatViewModel) ReceiveMessages(messageChan chan<- model.Message) {
	for {
		chatMessage, err := vm.chatService.ReceiveMessage()
		if err != nil {
			messageChan <- model.Message{Content: "Error receiving message: " + err.Error(), Sender: "System"}
			continue
		}
		chatMessageContent := chatMessage.GetChatMessage().GetMessage()
		senderUsername := chatMessage.GetFromUsername()
		message := model.Message{Content: chatMessageContent, Sender: senderUsername}
		vm.messages = append(vm.messages, message)
		messageChan <- message
	}
}

func (vm *ChatViewModel) AddMessage(message model.Message) {
	vm.messages = append(vm.messages, message)
}

func (vm *ChatViewModel) GetMessageCount() int {
	return len(vm.messages)
}

func (vm *ChatViewModel) GetMessageContent(index int) string {
	if index < 0 || index >= len(vm.messages) {
		return ""
	}
	msg := vm.messages[index]
	return fmt.Sprintf("%s", msg.Content)
}
