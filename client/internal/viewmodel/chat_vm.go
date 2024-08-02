package viewmodel

import (
	"client/internal/model"
	"client/internal/service"
	pb "client/resources/proto"
	"fmt"
	"sync"
)

type ChatViewModel struct {
	chatService             *service.ChatService
	chatterHandshakeService *service.ChatterHandshakeService
	messages                map[string][]model.Message
	CurrentChatter          string
	chatters                map[string]*model.Chatter
	onBack                  *func()
	messagesMutex           sync.RWMutex
	//
	commService *service.CommunicationService
}

func NewChatViewModel(commService *service.CommunicationService) *ChatViewModel {
	chatters := make(map[string]*model.Chatter)
	return &ChatViewModel{
		chatService:             service.NewChatService(commService),
		commService:             commService,
		messages:                make(map[string][]model.Message),
		chatters:                chatters,
		chatterHandshakeService: service.NewChatterHandshakeService(commService, chatters),
	}
}

func (vm *ChatViewModel) WaitForHandshakeMessages() {
	go func() {
		passiveKeyChan := vm.commService.GetPassiveKeyExchangeChannel()
		for {
			message := <-passiveKeyChan
			vm.handleHandshakeMessage(message)
		}
	}()
}

func (vm *ChatViewModel) handleHandshakeMessage(message *pb.Message) {
	vm.messagesMutex.Lock()
	defer vm.messagesMutex.Unlock()

	fromUsername := message.GetFromUsername()
	if _, exists := vm.chatters[fromUsername]; !exists {
		vm.chatters[fromUsername] = model.NewChatter(fromUsername)
	}

	vm.chatterHandshakeService.HandleReceiveHandshake(message)
}

func (vm *ChatViewModel) SetCurrentChat(username string) {
	vm.messagesMutex.Lock()
	defer vm.messagesMutex.Unlock()
	vm.CurrentChatter = username
	if _, exists := vm.messages[username]; !exists {
		vm.messages[username] = []model.Message{}
	}
	if _, exists := vm.chatters[username]; !exists {
		vm.chatters[username] = model.NewChatter(username)
		vm.chatterHandshakeService = service.NewChatterHandshakeService(vm.commService, vm.chatters)
		if err := vm.chatterHandshakeService.Handshake(username); err != nil {
			vm.AddMessage(model.Message{Content: "Error handshaking with user: " + err.Error(), Sender: "System"})
		}
	}
}

func (vm *ChatViewModel) SendMessage(content string, receiver string) {
	vm.messagesMutex.Lock()
	defer vm.messagesMutex.Unlock()

	chatter, exists := vm.chatters[receiver]
	if !exists {
		vm.AddMessage(model.Message{Content: "Error: Chatter not found", Sender: "System"})
		return
	}

	chatMessage := &pb.Message{
		Source:       pb.Message_CLIENT,
		FromUsername: &vm.commService.GetClient().Username,
		Packet: &pb.Message_ChatMessage{
			ChatMessage: &pb.ChatPacket{
				ToUsername: chatter.Username,
				Message:    chatter.Encrypt(content),
			},
		},
	}

	err := vm.chatService.SendMessage(chatMessage)
	if err != nil {
		vm.AddMessage(model.Message{Content: "Error sending message: " + err.Error(), Sender: "System"})
	} else {
		vm.AddMessage(model.Message{Content: content, Sender: "You", Receiver: chatter.Username})
	}
}

func (vm *ChatViewModel) ReceiveMessages(messageChan chan<- model.Message) {
	for {
		message, err := vm.chatService.ReceiveMessage()
		if err != nil {
			messageChan <- model.Message{Content: "Error receiving message: " + err.Error(), Sender: "System"}
			continue
		}

		chatMessage := message.GetChatMessage()
		if chatMessage == nil {
			continue
		}

		senderUsername := message.GetFromUsername()
		chatter, exists := vm.chatters[senderUsername]
		if !exists {
			messageChan <- model.Message{Content: "Error: Unknown sender", Sender: "System"}
			continue
		}

		decryptedContent := chatter.Decrypt(chatMessage.GetMessage())
		receivedMessage := model.Message{Content: decryptedContent, Sender: senderUsername, Receiver: senderUsername}
		vm.AddMessage(receivedMessage)
		messageChan <- receivedMessage
	}
}

func (vm *ChatViewModel) AddMessage(message model.Message) {
	vm.messagesMutex.Lock()
	defer vm.messagesMutex.Unlock()
	if message.Receiver == "" {
		message.Receiver = vm.CurrentChatter
	}
	vm.messages[message.Receiver] = append(vm.messages[message.Receiver], message)
}

func (vm *ChatViewModel) GetMessageCount() int {
	vm.messagesMutex.RLock()
	defer vm.messagesMutex.RUnlock()
	return len(vm.messages[vm.CurrentChatter])
}

func (vm *ChatViewModel) GetMessageContent(index int) string {
	vm.messagesMutex.RLock()
	defer vm.messagesMutex.RUnlock()
	if index < 0 || index >= len(vm.messages[vm.CurrentChatter]) {
		return ""
	}
	msg := vm.messages[vm.CurrentChatter][index]
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
