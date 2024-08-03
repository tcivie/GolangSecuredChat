package viewmodel

import (
	"client/internal/model"
	"client/internal/service"
	pb "client/resources/proto"
	"context"
	"fmt"
	"sync"
)

type ChatViewModel struct {
	chatService             *service.ChatService
	chatterHandshakeService *service.ChatterHandshakeService
	messages                *map[string][]model.Message
	CurrentChatter          string
	chatters                *map[string]*model.Chatter
	onBack                  *func()
	messagesMutex           sync.RWMutex
	commService             *service.CommunicationService
	messageChan             chan model.Message
	ctx                     context.Context
	cancelFunc              context.CancelFunc
}

func NewChatViewModel(commService *service.CommunicationService) *ChatViewModel {
	chatters := make(map[string]*model.Chatter)
	messages := make(map[string][]model.Message)
	return &ChatViewModel{
		chatService:             service.NewChatService(commService),
		commService:             commService,
		messages:                &messages,
		chatters:                &chatters,
		chatterHandshakeService: service.NewChatterHandshakeService(commService, &chatters),
		messageChan:             make(chan model.Message),
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
	if _, exists := (*vm.chatters)[fromUsername]; !exists {
		(*vm.chatters)[fromUsername] = model.NewChatter(fromUsername)
	}

	vm.chatterHandshakeService.HandleReceiveHandshake(message)
}

func (vm *ChatViewModel) SetCurrentChat(username string) {
	//vm.messagesMutex.Lock()
	//defer vm.messagesMutex.Unlock()
	vm.CurrentChatter = username
	if _, exists := (*vm.messages)[username]; !exists {
		(*vm.messages)[username] = []model.Message{}
	}
	if _, exists := (*vm.chatters)[username]; !exists {
		(*vm.chatters)[username] = model.NewChatter(username)
		if err := vm.chatterHandshakeService.Handshake(username); err != nil {
			vm.AddMessage(model.Message{Content: "Error handshaking with user: " + err.Error(), Sender: "System"})
		}
	}
}

func (vm *ChatViewModel) SendMessage(content string) {
	chatter, exists := (*vm.chatters)[vm.CurrentChatter]
	if !exists {
		vm.AddMessage(model.Message{Content: "Error: Chatter not found", Sender: "System"})
		return
	}
	fromUsername := vm.commService.GetUsername()
	chatMessage := &pb.Message{
		Source:       pb.Message_CLIENT,
		FromUsername: &fromUsername,
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

func (vm *ChatViewModel) StartReceivingMessages() {
	ctx, cancel := context.WithCancel(context.Background())
	vm.ctx = ctx
	vm.cancelFunc = cancel
	go vm.receiveMessages()
}

func (vm *ChatViewModel) StopReceivingMessages() {
	if vm.cancelFunc != nil {
		vm.cancelFunc()
	}
}

func (vm *ChatViewModel) GetMessageChan() <-chan model.Message {
	return vm.messageChan
}

func (vm *ChatViewModel) receiveMessages() {
	//defer vm.messagesMutex.Unlock()
	for {
		select {
		case <-vm.ctx.Done():
			if vm.messageChan != nil {
				close(vm.messageChan)
			}
			return
		default:
			message, err := vm.chatService.ReceiveMessage(vm.ctx)
			senderUsername := message.GetFromUsername()
			if err != nil {
				vm.messageChan <- model.Message{Content: "Error receiving message: " + err.Error(), Sender: "System", Receiver: vm.commService.GetUsername()}
				continue
			}
			if message == nil {
				return // context cancelled
			}

			chatMessage := message.GetChatMessage()
			if chatMessage == nil {
				continue
			}

			//vm.messagesMutex.Lock()
			chatter, exists := (*vm.chatters)[senderUsername]
			if !exists {
				vm.messageChan <- model.Message{Content: "Error: Unknown sender", Sender: "System", Receiver: vm.commService.GetUsername()}
				//vm.messagesMutex.Unlock()
				continue
			}

			decryptedContent := chatter.Decrypt(chatMessage.GetMessage())
			receivedMessage := model.Message{Content: decryptedContent, Sender: senderUsername, Receiver: vm.commService.GetUsername()}
			vm.messageChan <- receivedMessage
			//vm.messagesMutex.Unlock()
		}
	}
}

func (vm *ChatViewModel) AddMessage(message model.Message) {
	//vm.messagesMutex.Lock()
	//defer vm.messagesMutex.Unlock()
	if message.Receiver == "" {
		message.Receiver = vm.commService.GetUsername()
	}

	if message.Sender == "You" && message.Receiver == vm.CurrentChatter {
		(*vm.messages)[vm.CurrentChatter] = append((*vm.messages)[vm.CurrentChatter], message)
	} else if message.Sender == vm.CurrentChatter && message.Receiver == vm.commService.GetUsername() {
		(*vm.messages)[vm.CurrentChatter] = append((*vm.messages)[vm.CurrentChatter], message)
	} else {
		(*vm.messages)[message.Sender] = append((*vm.messages)[message.Sender], message)
	}
}

func (vm *ChatViewModel) GetMessageCount() int {
	//vm.messagesMutex.Lock()
	//defer vm.messagesMutex.Unlock()
	return len((*vm.messages)[vm.CurrentChatter])
}

func (vm *ChatViewModel) GetMessageContent(index int) string {
	//vm.messagesMutex.Lock()
	//defer vm.messagesMutex.Unlock()
	if index < 0 || index >= len((*vm.messages)[vm.CurrentChatter]) {
		return ""
	}
	msg := (*vm.messages)[vm.CurrentChatter][index]
	return fmt.Sprintf("%s: %s", msg.Sender, msg.Content)
}

func (vm *ChatViewModel) SetOnBack(callback func()) {
	vm.onBack = &callback
}

func (vm *ChatViewModel) Back() {
	vm.StopReceivingMessages()
	if vm.onBack != nil {
		(*vm.onBack)()
	}
}
