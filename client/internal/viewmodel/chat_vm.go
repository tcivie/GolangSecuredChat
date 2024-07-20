package viewmodel

import (
	"client/internal/model"
	"log"
)

type ChatViewModel struct {
	client   *model.Client
	messages []model.Message
}

func NewChatViewModel(client *model.Client) *ChatViewModel {
	return &ChatViewModel{
		client:   client,
		messages: []model.Message{},
	}
}

func (vm *ChatViewModel) SendMessage(content string) {
	err := vm.client.SendMessage(content)
	if err != nil {
		log.Printf("Error sending message: %v\n", err)
	}
}

func (vm *ChatViewModel) ReceiveMessages(messageChan chan<- model.Message) {
	for {
		content, err := vm.client.ReceiveMessage()
		if err != nil {
			log.Printf("Error receiving message: %v\n", err)
			return
		}
		message := model.Message{Content: content, Sender: "Server"}
		vm.messages = append(vm.messages, message)
		messageChan <- message
	}
}
