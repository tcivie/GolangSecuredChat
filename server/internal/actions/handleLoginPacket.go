package actions

import (
	"bytes"
	"crypto/rsa"
	"fmt"
	"google.golang.org/protobuf/proto"
	"net"
	"server/internal/db"
	"server/internal/util"
	pb "server/resources/proto"
)

type LoginMessageHandler struct {
	conn *net.TCPConn
	//
	loggingInUser string
	randomToken   []byte
}

func NewLoginMessageHandler(conn *net.TCPConn) *LoginMessageHandler {
	return &LoginMessageHandler{conn: conn}
}

func (h *LoginMessageHandler) handleMessage(message *pb.Message) error {
	var err error
	loginMessage := message.GetLoginMessage()
	if loginMessage == nil {
		return fmt.Errorf("unable to parse login message")
	}

	switch loginMessage.GetStatus() {
	case pb.LoginPacket_REQUEST_TO_LOGIN:
		var clientPublicKey *rsa.PublicKey
		var encryptedToken []byte

		fmt.Println("Received request to login")
		h.loggingInUser = message.GetFromUsername()

		// Pull from database the client's public key (Use the username hash to get the public key)
		hashedUsername := util.HashString(h.loggingInUser)
		database := db.GetDatabase()
		clientPublicKey, err = database.GetUserPubKey(hashedUsername)
		if err != nil {
			return fmt.Errorf("error getting public key from database: %v", err)
		}

		// Generate a random token with client's public key
		h.randomToken, err = util.GenerateRandomToken(512)
		if err != nil {
			return fmt.Errorf("error generating random token: %v", err)
		}

		// Encrypt the random token with the client's public key
		encryptedToken, err = util.EncodeUsingPubK(h.randomToken, clientPublicKey)
		if err != nil {
			return fmt.Errorf("error encrypting random token: %v", err)
		}

		// Send the encrypted token to the client
		loginReply := &pb.LoginPacket{
			Status: pb.LoginPacket_ENCRYPTED_TOKEN,
			Token:  encryptedToken,
		}
		err = h.sendLoginPacket(loginReply)
		break
	//case pb.LoginPacket_ENCRYPTED_TOKEN:
	//	fmt.Println("Received encrypted token")
	//	break
	case pb.LoginPacket_DECRYPTED_TOKEN:
		var loginReply *pb.LoginPacket

		fmt.Println("Received decrypted token")
		decodedToken := loginMessage.GetToken()

		if bytes.Compare(h.randomToken, decodedToken) == 0 {
			loginReply = &pb.LoginPacket{
				Status: pb.LoginPacket_LOGIN_SUCCESS,
			}
		} else {
			loginReply = &pb.LoginPacket{
				Status: pb.LoginPacket_LOGIN_FAILED,
			}
		}
		err = h.sendLoginPacket(loginReply)
		break
	//case pb.LoginPacket_LOGIN_REPLY:
	//	fmt.Println("Received login reply")
	//	break
	default:
		return fmt.Errorf("unknown login message status")
	}
	return err
}

func (h *LoginMessageHandler) sendLoginPacket(reply *pb.LoginPacket) error {
	message := &pb.Message{
		Source: pb.Message_SERVER,
		Packet: &pb.Message_LoginMessage{
			LoginMessage: reply,
		},
	}

	data, err := proto.Marshal(message)
	if err != nil {
		return fmt.Errorf("error marshalling message: %v", err)
	}
	_, err = h.conn.Write(data)
	if err != nil {
		return fmt.Errorf("error sending message: %v", err)
	}

	return nil
}
