package util

import (
	"encoding/binary"
	"fmt"
	"google.golang.org/protobuf/proto"
	"io"
	"log"
	"net"
	pb "server/resources/proto"
)

func SendMessage(conn net.Conn, message *pb.Message) error {
	data, err := proto.Marshal(message)
	if err != nil {
		return fmt.Errorf("error marshalling message: %v", err)
	}

	// Write the length of the message
	log.Println("SendMessage\t len: ", len(data))
	err = binary.Write(conn, binary.BigEndian, uint32(len(data)))
	if err != nil {
		return fmt.Errorf("error writing message length: %v", err)
	}

	// Write the message itself
	//log.Println("SendMessage data: ", data)
	_, err = conn.Write(data)
	if err != nil {
		return fmt.Errorf("error writing message: %v", err)
	}

	return nil
}

// ReadMessage reads a length-prefixed message from the connection
func ReadMessage(conn net.Conn) (*pb.Message, error) {
	// Read the message length
	var length uint32
	err := binary.Read(conn, binary.BigEndian, &length)
	log.Println("ReadMessage\t len: ", length)
	if err != nil {
		return nil, fmt.Errorf("error reading message length: %v", err)
	}

	// Read the message data
	data := make([]byte, length)
	_, err = io.ReadFull(conn, data)
	//log.Println("ReadMessage data: ", data)
	if err != nil {
		return nil, fmt.Errorf("error reading message data: %v", err)
	}

	// Unmarshal the message
	message := &pb.Message{}
	if err := proto.Unmarshal(data, message); err != nil {
		return nil, fmt.Errorf("error unmarshalling message: %v", err)
	}

	return message, nil
}
