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
	Client  *model.Client
	Chatter *model.Chatter
}

func NewChatterHandshakeService(client *model.Client, chatter *model.Chatter) *ChatterHandshakeService {
	return &ChatterHandshakeService{
		Client:  client,
		Chatter: chatter,
	}
}

func (s *ChatterHandshakeService) Handshake() error {
	// Request Chatter's public key from server
	publicKeyRequest := &pb.ExchangeKeyPacket{
		Status:     pb.ExchangeKeyPacket_REQUEST_FOR_USER_PUBLIC_KEY,
		ToUsername: &s.Chatter.Username,
	}
	err := s.sendHandshakeMessage(publicKeyRequest)
	if err != nil {
		return err
	}

	// Receive Chatter's public key
	response, err := s.Client.GetMessage()
	if err != nil {
		return err
	}
	publicKeyMessage := response.GetExchangeKeyMessage()
	if publicKeyMessage == nil || publicKeyMessage.GetStatus() != pb.ExchangeKeyPacket_PUB_KEY_FROM_SERVER {
		return errors.New("invalid public key response")
	}

	// Construct the RSA public key
	pubkeyInt := new(big.Int).SetBytes(publicKeyMessage.GetKey())
	sshPubKey := &rsa.PublicKey{
		N: pubkeyInt,
		E: 65537, // Commonly used public exponent
	}
	s.Chatter.SetPublicKey(sshPubKey)

	// Encrypt our public key with Chatter's public key
	encryptedPublicKey, err := s.Chatter.EncryptWithPublicKey(s.Client.GetPubKey().N.Bytes())
	if err != nil {
		return err
	}

	// Send our public key to Chatter
	publicKeyResponse := &pb.ExchangeKeyPacket{
		Status:           pb.ExchangeKeyPacket_MESSAGE_WITH_PUB_KEY,
		EncryptedMessage: encryptedPublicKey,
	}
	err = s.sendHandshakeMessage(publicKeyResponse)
	if err != nil {
		return err
	}

	// Receive Chatter's symmetric key
	responseWithSymKey, err := s.Client.GetMessage()
	if err != nil {
		return err
	}
	messageWithSymKey := responseWithSymKey.GetExchangeKeyMessage()
	if messageWithSymKey == nil || messageWithSymKey.GetStatus() != pb.ExchangeKeyPacket_REPLY_WITH_SYM_KEY {
		return errors.New("invalid public key response")
	}

	// Decrypt the AES key with our private key
	decryptedAESKey, err := s.Client.DecryptMessageWithPrivateKey(messageWithSymKey.GetEncryptedMessage())
	if err != nil {
		return err
	}
	s.Chatter.SetAES256Key(decryptedAESKey)
	return nil
}

func (s *ChatterHandshakeService) HandleReciveHandshake(message *pb.Message) {
	exchangeKeyMessage := message.GetExchangeKeyMessage()
	if exchangeKeyMessage == nil {
		return
	}
	switch exchangeKeyMessage.GetStatus() {
	case pb.ExchangeKeyPacket_MESSAGE_WITH_PUB_KEY:
		// Decrypt the message with our private key
		decryptedMessage, err := s.Client.DecryptMessageWithPrivateKey(exchangeKeyMessage.GetEncryptedMessage())
		if err != nil {
			fmt.Println("Error decrypting message: ", err)
			return
		}

		// Generate a random AES key and encrypt it with the Chatter's public key
		randomAESKey := make([]byte, 32)
		_, err = rand.Read(randomAESKey)
		if err != nil {
			fmt.Println("Error generating random AES key: ", err)
			return
		}
		encryptedAESKey, err := s.Chatter.EncryptWithPublicKey(s.Chatter.aes256Key)
		// TODO: Send the encrypted AES key to the Chatter and save it in the Chatter struct and handle the rest of the cases
	}
}

func (s *ChatterHandshakeService) sendHandshakeMessage(message *pb.ExchangeKeyPacket) error {
	handShakeMessage := &pb.Message{
		Source:       pb.Message_CLIENT,
		FromUsername: &s.Client.Username,
		Packet: &pb.Message_ExchangeKeyMessage{
			ExchangeKeyMessage: message,
		},
	}
	return s.Client.SendMessage(handShakeMessage)
}
