package service

import (
	"client/internal/model"
	pb "client/resources/proto"
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"fmt"
	"math/big"
)

type ChatterHandshakeService struct {
	commService *CommunicationService
	Chatters    map[string]*model.Chatter
}

func NewChatterHandshakeService(commService *CommunicationService, chatters map[string]*model.Chatter) *ChatterHandshakeService {
	return &ChatterHandshakeService{
		commService: commService,
		Chatters:    chatters,
	}
}

func (s *ChatterHandshakeService) Handshake(username string) error {
	// Request Chatter's public key from server
	publicKeyRequest := &pb.ExchangeKeyPacket{
		Status:     pb.ExchangeKeyPacket_REQUEST_FOR_USER_PUBLIC_KEY,
		ToUsername: &s.Chatters[username].Username,
	}
	err := s.sendHandshakeMessage(publicKeyRequest)
	if err != nil {
		fmt.Println("Error sending public key request: ", err)
		return err
	}

	// Receive Chatter's public key
	keyChan := s.commService.GetKeyExchangeChannel()
	publicKeyMessage := <-keyChan
	if publicKeyMessage == nil || publicKeyMessage.GetStatus() != pb.ExchangeKeyPacket_PUB_KEY_FROM_SERVER {
		return errors.New("invalid public key response")
	}

	// Construct the RSA public key
	pubkeyInt := new(big.Int).SetBytes(publicKeyMessage.GetKey())
	sshPubKey := &rsa.PublicKey{
		N: pubkeyInt,
		E: 65537, // Commonly used public exponent
	}
	s.Chatters[username].SetPublicKey(sshPubKey)

	// Encrypt our username with Chatter's public key
	encryptedUsername, err := s.Chatters[username].EncryptWithPublicKey([]byte(s.commService.GetClient().Username))
	if err != nil {
		fmt.Println("Error encrypting public key: ", err)
		return err
	}

	// Send our username to Chatter
	publicKeyResponse := &pb.ExchangeKeyPacket{
		Status:           pb.ExchangeKeyPacket_REQ_FOR_SYM_KEY,
		EncryptedMessage: encryptedUsername,
		ToUsername:       &s.Chatters[username].Username,
	}
	err = s.sendHandshakeMessage(publicKeyResponse)
	if err != nil {
		fmt.Println("Error sending public key response: ", err)
		return err
	}

	// Receive Chatter's symmetric key
	messageWithSymKey := <-keyChan
	if messageWithSymKey == nil || messageWithSymKey.GetStatus() != pb.ExchangeKeyPacket_REPLY_WITH_SYM_KEY {
		return errors.New("invalid public key response")
	}

	// Decrypt the AES key with our private key
	decryptedAESKey, err := s.commService.GetClient().DecryptMessageWithPrivateKey(messageWithSymKey.GetEncryptedMessage())
	if err != nil {
		fmt.Println("Error decrypting AES key: ", err)
		return err
	}
	s.Chatters[username].SetAES256Key(decryptedAESKey)
	return nil
}

func (s *ChatterHandshakeService) HandleReceiveHandshake(message *pb.Message) {
	var response *pb.ExchangeKeyPacket
	fromUsername := message.GetFromUsername()
	exchangeKeyMessage := message.GetExchangeKeyMessage()

	switch exchangeKeyMessage.GetStatus() {
	case pb.ExchangeKeyPacket_REQ_FOR_SYM_KEY:
		// Decrypt the exchangeKeyMessage with our private key
		decryptedMessage, err := s.commService.GetClient().DecryptMessageWithPrivateKey(exchangeKeyMessage.GetEncryptedMessage())
		if err != nil {
			fmt.Println("Error decrypting exchangeKeyMessage: ", err)
			return
		}

		username := string(decryptedMessage[:])

		if username != s.commService.GetClient().Username {
			fmt.Println("Invalid username in request for symmetric key")
			response = &pb.ExchangeKeyPacket{
				Status:     pb.ExchangeKeyPacket_ERROR,
				ToUsername: &s.Chatters[fromUsername].Username,
			}
			break
		}

		// Verify with the server that the public key is valid (Ask for the public key from the server)
		response = &pb.ExchangeKeyPacket{
			Status:     pb.ExchangeKeyPacket_REQUEST_FOR_USER_PUBLIC_KEY_PASSIVE,
			ToUsername: &username,
		}
	case pb.ExchangeKeyPacket_PUB_KEY_FROM_SERVER_PASSIVE:
		// Update the chatter with the public key
		if exchangeKeyMessage.GetToUsername() != s.Chatters[fromUsername].Username {
			fmt.Println("Invalid username in public key from server")
			// Send error to the chatter
			response = &pb.ExchangeKeyPacket{
				Status:     pb.ExchangeKeyPacket_ERROR,
				ToUsername: &s.Chatters[fromUsername].Username,
			}
			break
		}
		s.Chatters[fromUsername].SetPublicKey(&rsa.PublicKey{
			N: new(big.Int).SetBytes(exchangeKeyMessage.GetKey()),
			E: 65537,
		})
		// Generate a random AES key and encrypt it with the Chatter's public key
		randomAESKey := make([]byte, 32)
		_, err := rand.Read(randomAESKey)
		if err != nil {
			fmt.Println("Error generating random AES key: ", err)
			return
		}
		s.Chatters[fromUsername].SetAES256Key(randomAESKey)
		encryptedAESKey, err := s.Chatters[fromUsername].EncryptWithPublicKey(randomAESKey)
		if err != nil {
			fmt.Println("Error encrypting AES key: ", err)
			return
		}

		// Send the encrypted AES key to the Chatter
		response = &pb.ExchangeKeyPacket{
			Status:           pb.ExchangeKeyPacket_REPLY_WITH_SYM_KEY,
			EncryptedMessage: encryptedAESKey,
			ToUsername:       &s.Chatters[fromUsername].Username,
		}
	}

	if response != nil {
		err := s.sendHandshakeMessage(response)
		if err != nil {
			fmt.Println("Error sending handshake exchangeKeyMessage: ", err)
		}
	}
}

func (s *ChatterHandshakeService) sendHandshakeMessage(message *pb.ExchangeKeyPacket) error {
	handShakeMessage := &pb.Message{
		Source:       pb.Message_CLIENT,
		FromUsername: &s.commService.GetClient().Username,
		Packet: &pb.Message_ExchangeKeyMessage{
			ExchangeKeyMessage: message,
		},
	}
	return s.commService.SendMessage(handShakeMessage)
}
