package view

import (
	"bufio"
	"fmt"
	"os"

	"client/internal/model"
	"client/internal/viewmodel"
)

type ChatView struct {
	viewModel *viewmodel.ChatViewModel
}

func NewChatView(vm *viewmodel.ChatViewModel) *ChatView {
	return &ChatView{viewModel: vm}
}

func (v *ChatView) Run() {
	fmt.Println("Connected to chat server. Start typing messages:")

	messageChan := make(chan model.Message)
	go v.viewModel.ReceiveMessages(messageChan)

	go func() {
		for message := range messageChan {
			fmt.Printf("Received: %s", message.Content)
		}
	}()

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		message := scanner.Text()
		v.viewModel.SendMessage(message)
	}
}
