package viewmodel

import (
	"client/internal/model"
	"client/internal/service"
	"fmt"
)

type ChatViewModel struct {
	service  *service.ChatService
	messages []model.Message
}

func NewChatViewModel(service *service.ChatService) *ChatViewModel {
	return &ChatViewModel{
		service:  service,
		messages: []model.Message{},
	}
}

func (vm *ChatViewModel) SendMessage(content string) {
	err := vm.service.SendMessage(content)
	if err != nil {
		vm.AddMessage(model.Message{Content: "Error sending message: " + err.Error(), Sender: "System"})
	} else {
		vm.AddMessage(model.Message{Content: content, Sender: "You"})
	}
}

func (vm *ChatViewModel) ReceiveMessages(messageChan chan<- model.Message) {
	for {
		content, err := vm.service.ReceiveMessage()
		if err != nil {
			messageChan <- model.Message{Content: "Error receiving message: " + err.Error(), Sender: "System"}
			continue
		}
		message := model.Message{Content: content, Sender: "Server"}
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
	return fmt.Sprintf("%s: %s", msg.Sender, msg.Content)
}
