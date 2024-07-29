package viewmodel

import (
	"client/internal/model"
	"client/internal/service"
	pb "client/resources/proto"
	"fmt"
	"sync"
)

type ChatViewModel struct {
	chatService   *service.ChatService
	messages      map[string][]model.Message
	currentChat   string
	messagesMutex sync.RWMutex
	onBack        *func()
}

func NewChatViewModel(service *service.ChatService) *ChatViewModel {
	return &ChatViewModel{
		chatService: service,
		messages:    make(map[string][]model.Message),
	}
}

func (vm *ChatViewModel) SetCurrentChat(username string) {
	vm.messagesMutex.Lock()
	defer vm.messagesMutex.Unlock()
	vm.currentChat = username
	if _, exists := vm.messages[username]; !exists {
		vm.messages[username] = []model.Message{}
	}
}

func (vm *ChatViewModel) SendMessage(content string) {
	chatMessage := &pb.Message{
		Source:       pb.Message_CLIENT,
		FromUsername: &vm.chatService.Client.Username,
		Packet: &pb.Message_ChatMessage{
			ChatMessage: &pb.ChatPacket{
				ToUsername: vm.currentChat,
				Message:    content,
			},
		},
	}
	err := vm.chatService.SendMessage(chatMessage)
	if err != nil {
		vm.AddMessage(model.Message{Content: "Error sending message: " + err.Error(), Sender: "System"})
	} else {
		vm.AddMessage(model.Message{Content: content, Sender: "You", Receiver: vm.currentChat})
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
		message := model.Message{Content: chatMessageContent, Sender: senderUsername, Receiver: vm.currentChat}
		vm.AddMessage(message)
		messageChan <- message
	}
}

func (vm *ChatViewModel) AddMessage(message model.Message) {
	vm.messagesMutex.Lock()
	defer vm.messagesMutex.Unlock()
	if message.Receiver == "" {
		message.Receiver = vm.currentChat
	}
	vm.messages[message.Receiver] = append(vm.messages[message.Receiver], message)
}

func (vm *ChatViewModel) GetMessageCount() int {
	vm.messagesMutex.RLock()
	defer vm.messagesMutex.RUnlock()
	return len(vm.messages[vm.currentChat])
}

func (vm *ChatViewModel) GetMessageContent(index int) string {
	vm.messagesMutex.RLock()
	defer vm.messagesMutex.RUnlock()
	if index < 0 || index >= len(vm.messages[vm.currentChat]) {
		return ""
	}
	msg := vm.messages[vm.currentChat][index]
	return fmt.Sprintf("%s: %s", msg.Sender, msg.Content)
}

func (vm *ChatViewModel) SetOnBack(callback func()) {
	vm.onBack = &callback
}

func (vm *ChatViewModel) Back() {
	if vm.onBack != nil {
		(*vm.onBack)()
	}
}
