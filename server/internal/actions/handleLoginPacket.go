package actions

import (
	"bytes"
	"crypto/rsa"
	"crypto/sha256"
	"fmt"
	"net"
	"server/internal/db"
	"server/internal/util"
	pb "server/resources/proto"
)

type LoginMessageHandler struct {
	conn net.Conn
	//
	loggingInUser       string
	randomToken         []byte
	listOfLoggedInUsers *map[string]net.Conn
}

func NewLoginMessageHandler(conn net.Conn, listOfLoggedInUsers *map[string]net.Conn) *LoginMessageHandler {
	return &LoginMessageHandler{conn: conn, listOfLoggedInUsers: listOfLoggedInUsers}
}

func (h *LoginMessageHandler) handleMessage(message *pb.Message) error {
	var err error
	var loginReply *pb.LoginPacket
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
			loginReply = &pb.LoginPacket{
				Status: pb.LoginPacket_LOGIN_FAILED,
			}

			return fmt.Errorf("error getting public key from database: %v", err)
		}

		maxTokenLength := clientPublicKey.Size() - 2*sha256.Size - 2
		// Generate a random token with client's public key
		h.randomToken, err = util.GenerateRandomToken(maxTokenLength)
		if err != nil {
			loginReply = &pb.LoginPacket{
				Status: pb.LoginPacket_LOGIN_FAILED,
			}
			fmt.Println("error generating random token: %v", err)
			break
		}

		// Encrypt the random token with the client's public key
		encryptedToken, err = util.EncodeUsingPubK(h.randomToken, clientPublicKey)
		if err != nil {
			loginReply = &pb.LoginPacket{
				Status: pb.LoginPacket_LOGIN_FAILED,
			}
			fmt.Println("error encrypting random token: %v", err)
			break
		}

		// Send the encrypted token to the client
		loginReply = &pb.LoginPacket{
			Status: pb.LoginPacket_ENCRYPTED_TOKEN,
			Token:  encryptedToken,
		}
		break
	//case pb.LoginPacket_ENCRYPTED_TOKEN:
	//	fmt.Println("Received encrypted token")
	//	break
	case pb.LoginPacket_DECRYPTED_TOKEN:
		fmt.Println("Received decrypted token")
		decodedToken := loginMessage.GetToken()

		if bytes.Compare(h.randomToken, decodedToken) == 0 {
			fmt.Println("Login successful")
			(*h.listOfLoggedInUsers)[h.loggingInUser] = h.conn
			loginReply = &pb.LoginPacket{
				Status: pb.LoginPacket_LOGIN_SUCCESS,
			}
		} else {
			fmt.Println("Login failed")
			loginReply = &pb.LoginPacket{
				Status: pb.LoginPacket_LOGIN_FAILED,
			}
		}
		break
	default:
		return fmt.Errorf("unknown login message status")
	}
	_ = h.sendLoginPacket(loginReply)
	return err
}

func (h *LoginMessageHandler) sendLoginPacket(reply *pb.LoginPacket) error {
	message := &pb.Message{
		Source: pb.Message_SERVER,
		Packet: &pb.Message_LoginMessage{
			LoginMessage: reply,
		},
	}

	return util.SendMessage(h.conn, message)
}
