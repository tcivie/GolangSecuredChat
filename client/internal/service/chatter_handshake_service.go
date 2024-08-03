package service

import (
	"client/internal/model"
	"client/internal/utils"
	pb "client/resources/proto"
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"fmt"
	"math/big"
)

type ChatterHandshakeService struct {
	commService *CommunicationService
	Chatters    *map[string]*model.Chatter
}

func NewChatterHandshakeService(commService *CommunicationService, chatters *map[string]*model.Chatter) *ChatterHandshakeService {
	return &ChatterHandshakeService{
		commService: commService,
		Chatters:    chatters,
	}
}

func (s *ChatterHandshakeService) Handshake(username string) error {
	// Request Chatter's public key from server
	publicKeyRequest := &pb.ExchangeKeyPacket{
		Status:     pb.ExchangeKeyPacket_REQUEST_FOR_USER_PUBLIC_KEY,
		ToUsername: &(*s.Chatters)[username].Username,
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
	fmt.Printf("Setting public key (%s) for %s\n", utils.DebugPrintPublicKey(sshPubKey), username)
	(*s.Chatters)[username].SetPublicKey(sshPubKey)

	// Encrypt our username with Chatter's public key
	fmt.Printf("Encrypting with public key (%s) for %s\n", utils.DebugPrintPublicKey(sshPubKey), username)
	encryptedUsername, err := (*s.Chatters)[username].EncryptWithPublicKey([]byte(s.commService.GetUsername()))
	if err != nil {
		fmt.Println("Error encrypting public key: ", err)
		return err
	}

	// Send our username to Chatter
	publicKeyResponse := &pb.ExchangeKeyPacket{
		Status:           pb.ExchangeKeyPacket_REQ_FOR_SYM_KEY,
		EncryptedMessage: encryptedUsername,
		ToUsername:       &(*s.Chatters)[username].Username,
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
	fmt.Printf("Decrpyting with private key for this public (%s) for %s\n", utils.DebugPrintPublicKey(s.commService.GetClient().GetPubKey()), s.commService.GetUsername())
	decryptedAESKey, err := s.commService.GetClient().DecryptMessageWithPrivateKey(messageWithSymKey.GetEncryptedMessage())
	if err != nil {
		fmt.Println("Error decrypting AES key: ", err)
		return err
	}
	(*s.Chatters)[username].SetAES256Key(decryptedAESKey)
	return nil
}

func (s *ChatterHandshakeService) HandleReceiveHandshake(message *pb.Message) {
	var response *pb.ExchangeKeyPacket
	fromUsername := message.GetFromUsername()
	exchangeKeyMessage := message.GetExchangeKeyMessage()
	destinationUsername := exchangeKeyMessage.GetToUsername()

	switch exchangeKeyMessage.GetStatus() {
	case pb.ExchangeKeyPacket_REQ_FOR_SYM_KEY:
		// Decrypt the exchangeKeyMessage with our private key
		fmt.Printf("Decrpyting with private key for this public (%s) for %s\n", utils.DebugPrintPublicKey(s.commService.GetClient().GetPubKey()), s.commService.GetUsername())
		decryptedMessage, err := s.commService.GetClient().DecryptMessageWithPrivateKey(exchangeKeyMessage.GetEncryptedMessage())
		if err != nil {
			fmt.Println("Error decrypting exchangeKeyMessage: ", err)
			return
		}

		username := string(decryptedMessage[:])

		if username != fromUsername {
			fmt.Println("Invalid username in request for symmetric key")
			response = &pb.ExchangeKeyPacket{
				Status:     pb.ExchangeKeyPacket_ERROR,
				ToUsername: &(*s.Chatters)[fromUsername].Username,
			}
			break
		}

		// Check if the Chatter exists (if not, create it)
		if _, exists := (*s.Chatters)[fromUsername]; !exists {
			(*s.Chatters)[fromUsername] = model.NewChatter(fromUsername)
		}

		// Verify with the server that the public key is valid (Ask for the public key from the server)
		response = &pb.ExchangeKeyPacket{
			Status:     pb.ExchangeKeyPacket_REQUEST_FOR_USER_PUBLIC_KEY_PASSIVE,
			ToUsername: &username,
		}
	case pb.ExchangeKeyPacket_PUB_KEY_FROM_SERVER_PASSIVE:
		// Update the chatter with the public key
		if message.GetSource() != pb.Message_SERVER {
			fmt.Println("Key must be from server")
			// Send error to the chatter
			response = &pb.ExchangeKeyPacket{
				Status:     pb.ExchangeKeyPacket_ERROR,
				ToUsername: &(*s.Chatters)[destinationUsername].Username,
			}
			break
		}
		(*s.Chatters)[destinationUsername].SetPublicKey(&rsa.PublicKey{
			N: new(big.Int).SetBytes(exchangeKeyMessage.GetKey()),
			E: 65537,
		})
		fmt.Printf("Setting public key (%s) for %s\n", utils.DebugPrintPublicKey((*s.Chatters)[destinationUsername].GetPubKey()), destinationUsername)

		// Generate a random AES key and encrypt it with the Chatter's public key
		randomAESKey := make([]byte, 32)
		_, err := rand.Read(randomAESKey)
		if err != nil {
			fmt.Println("Error generating random AES key: ", err)
			return
		}
		(*s.Chatters)[destinationUsername].SetAES256Key(randomAESKey)
		fmt.Printf("Encrypting with public key (%s) for %s\n", utils.DebugPrintPublicKey((*s.Chatters)[destinationUsername].GetPubKey()), destinationUsername)
		encryptedAESKey, err := (*s.Chatters)[destinationUsername].EncryptWithPublicKey(randomAESKey)
		if err != nil {
			fmt.Println("Error encrypting AES key: ", err)
			return
		}

		// Send the encrypted AES key to the Chatter
		response = &pb.ExchangeKeyPacket{
			Status:           pb.ExchangeKeyPacket_REPLY_WITH_SYM_KEY,
			EncryptedMessage: encryptedAESKey,
			ToUsername:       &(*s.Chatters)[destinationUsername].Username,
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
	fromUsername := s.commService.GetUsername()
	handShakeMessage := &pb.Message{
		Source:       pb.Message_CLIENT,
		FromUsername: &fromUsername,
		Packet: &pb.Message_ExchangeKeyMessage{
			ExchangeKeyMessage: message,
		},
	}
	return s.commService.SendMessage(handShakeMessage)
}
