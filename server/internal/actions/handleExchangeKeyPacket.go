package actions

import (
	"fmt"
	"net"
	"server/internal/db"
	"server/internal/util"
	pb "server/resources/proto"
)

type ExchangeKeyPacket struct {
	listOfLoggedInUsers *map[string]net.Conn
}

func NewExchangeKeyPacket(listOfLoggedInUsers *map[string]net.Conn) *ExchangeKeyPacket {
	return &ExchangeKeyPacket{listOfLoggedInUsers: listOfLoggedInUsers}
}

func (ekp *ExchangeKeyPacket) handleMessage(message *pb.Message) error {
	var exchangeKeyReply *pb.ExchangeKeyPacket
	var destinationConn net.Conn
	exchangeKeyMessage := message.GetExchangeKeyMessage()
	if exchangeKeyMessage == nil {
		return fmt.Errorf("unable to parse exchange key message")
	}
	sourceUser := message.GetFromUsername()

	switch exchangeKeyMessage.GetStatus() {
	case pb.ExchangeKeyPacket_REQUEST_FOR_USER_PUBLIC_KEY:
		fmt.Println("Received request for user public key")
		// Pull from database the client's public key (Use the username hash to get the public key)
		hashedUsername := util.HashString(message.GetFromUsername())
		clientPublicKey, err := db.GetDatabase().GetUserPubKey(hashedUsername)
		if err != nil {
			exchangeKeyReply = &pb.ExchangeKeyPacket{
				Status: pb.ExchangeKeyPacket_ERROR,
			}
			fmt.Printf("error getting public key from database: %v\n", err)
			destinationConn = (*ekp.listOfLoggedInUsers)[message.GetFromUsername()] // Return to sender
			break
		}

		exchangeKeyReply = &pb.ExchangeKeyPacket{
			Status: pb.ExchangeKeyPacket_PUB_KEY_FROM_SERVER,
			Key:    clientPublicKey.N.Bytes(),
		}
		destinationConn = (*ekp.listOfLoggedInUsers)[message.GetFromUsername()] // Return to sender
		break
	case pb.ExchangeKeyPacket_REQUEST_FOR_USER_PUBLIC_KEY_PASSIVE:
		fmt.Println("Received request for user public key")
		// Pull from database the client's public key (Use the username hash to get the public key)
		hashedUsername := util.HashString(message.GetFromUsername())
		clientPublicKey, err := db.GetDatabase().GetUserPubKey(hashedUsername)
		if err != nil {
			exchangeKeyReply = &pb.ExchangeKeyPacket{
				Status: pb.ExchangeKeyPacket_ERROR,
			}
			fmt.Printf("error getting public key from database: %v\n", err)
			destinationConn = (*ekp.listOfLoggedInUsers)[message.GetFromUsername()] // Return to sender
			break
		}

		exchangeKeyReply = &pb.ExchangeKeyPacket{
			Status: pb.ExchangeKeyPacket_PUB_KEY_FROM_SERVER_PASSIVE,
			Key:    clientPublicKey.N.Bytes(),
		}
		destinationConn = (*ekp.listOfLoggedInUsers)[message.GetFromUsername()] // Return to sender
		break
	case pb.ExchangeKeyPacket_REQ_FOR_SYM_KEY:
		fmt.Println("Received request for symmetric key")
		// Forward the message as is to the recipient
		exchangeKeyReply = exchangeKeyMessage
		destinationConn = (*ekp.listOfLoggedInUsers)[exchangeKeyMessage.GetToUsername()]
		break
	case pb.ExchangeKeyPacket_REPLY_WITH_SYM_KEY:
		fmt.Println("Received reply with symmetric key")
		// Forward the message as is to the recipient
		exchangeKeyReply = exchangeKeyMessage
		destinationConn = (*ekp.listOfLoggedInUsers)[exchangeKeyMessage.GetToUsername()]
		break
	default:
		exchangeKeyReply = &pb.ExchangeKeyPacket{
			Status: pb.ExchangeKeyPacket_ERROR,
		}
		destinationConn = (*ekp.listOfLoggedInUsers)[message.GetFromUsername()]
		fmt.Printf("invalid exchange key message status")
	}

	if destinationConn == nil {
		return fmt.Errorf("destination connection not found")
	}
	return ekp.sendExchangeKeyMessage(exchangeKeyReply, destinationConn, sourceUser)
}

func (ekp *ExchangeKeyPacket) sendExchangeKeyMessage(exchangeKeyMessage *pb.ExchangeKeyPacket, destination net.Conn, sourceUser string) error {
	message := &pb.Message{
		Source:       pb.Message_SERVER,
		FromUsername: &sourceUser,
		Packet: &pb.Message_ExchangeKeyMessage{
			ExchangeKeyMessage: exchangeKeyMessage,
		},
	}

	return util.SendMessage(destination, message)
}
