package service

import (
	"client/internal/model"
	pb "client/resources/proto"
	"fmt"
	"log"
	"sync"
)

type CommunicationService struct {
	client *model.Client
	// Channels for different types of messages
	loginChan      chan *pb.LoginPacket
	registerChan   chan *pb.RegisterPacket
	chatChan       chan *pb.Message
	userListChan   chan *pb.UserListPacket
	keyChan        chan *pb.ExchangeKeyPacket
	passiveKeyChan chan *pb.Message
	errorChan      chan error
	// Mutex to protect concurrent access
	mu sync.Mutex
}

func NewCommunicationService(client *model.Client) *CommunicationService {
	cs := &CommunicationService{
		client:         client,
		loginChan:      make(chan *pb.LoginPacket),
		registerChan:   make(chan *pb.RegisterPacket),
		chatChan:       make(chan *pb.Message),
		userListChan:   make(chan *pb.UserListPacket),
		keyChan:        make(chan *pb.ExchangeKeyPacket),
		passiveKeyChan: make(chan *pb.Message),
		errorChan:      make(chan error),
	}

	go cs.handleMessages()

	return cs
}

func (cs *CommunicationService) handleMessages() {
	for {
		message, err := cs.client.GetMessage()
		if err != nil {
			log.Printf("Error receiving message: %v", err)
			cs.errorChan <- fmt.Errorf("communication error: %v", err)
			if err.Error() == "connection closed by server" {
				// Handle disconnection
				log.Println("Disconnected from server")
				// You might want to implement a reconnection mechanism here
				return
			}
			continue
		}

		switch msg := message.Packet.(type) {
		case *pb.Message_LoginMessage:
			cs.loginChan <- msg.LoginMessage
		case *pb.Message_RegisterMessage:
			cs.registerChan <- msg.RegisterMessage
		case *pb.Message_ChatMessage:
			cs.chatChan <- message
		case *pb.Message_UserListMessage:
			cs.userListChan <- msg.UserListMessage
		case *pb.Message_ExchangeKeyMessage:
			switch message.GetPacket().(*pb.Message_ExchangeKeyMessage).ExchangeKeyMessage.GetStatus() {
			case pb.ExchangeKeyPacket_REQUEST_FOR_USER_PUBLIC_KEY,
				pb.ExchangeKeyPacket_REPLY_WITH_SYM_KEY,
				pb.ExchangeKeyPacket_PUB_KEY_FROM_SERVER,
				pb.ExchangeKeyPacket_ERROR:
				cs.keyChan <- msg.ExchangeKeyMessage
			case pb.ExchangeKeyPacket_REQ_FOR_SYM_KEY,
				pb.ExchangeKeyPacket_PUB_KEY_FROM_SERVER_PASSIVE:
				cs.passiveKeyChan <- message
			}

		default:
			log.Printf("Received unknown message type: %T", msg)
		}
	}
}

func (cs *CommunicationService) SendMessage(message *pb.Message) error {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	return cs.client.SendMessage(message)
}

func (cs *CommunicationService) GetLoginChannel() <-chan *pb.LoginPacket {
	return cs.loginChan
}

func (cs *CommunicationService) GetRegisterChannel() <-chan *pb.RegisterPacket {
	return cs.registerChan
}

func (cs *CommunicationService) GetChatChannel() <-chan *pb.Message {
	return cs.chatChan
}

func (cs *CommunicationService) GetUserListChannel() <-chan *pb.UserListPacket {
	return cs.userListChan
}

func (cs *CommunicationService) GetKeyExchangeChannel() <-chan *pb.ExchangeKeyPacket {
	return cs.keyChan
}

func (cs *CommunicationService) GetPassiveKeyExchangeChannel() <-chan *pb.Message {
	return cs.passiveKeyChan
}

func (cs *CommunicationService) GetClient() *model.Client {
	return cs.client
}

func (cs *CommunicationService) GetUsername() string {
	return cs.client.Username
}

func (cs *CommunicationService) SetClientUsername(username string) {
	cs.client.Username = username
}

func (cs *CommunicationService) GetErrorChannel() <-chan error {
	return cs.errorChan
}
