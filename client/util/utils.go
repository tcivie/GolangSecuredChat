package util

import (
	"client/resources/proto"
	"crypto/tls"
	"google.golang.org/protobuf/proto"
)

func GetMessage(conn *tls.Conn, message *pb.Message) (*pb.Message, error) {
	buffer := make([]byte, 1024*4)
	n, err := conn.Read(buffer)
	if err != nil {
		return nil, err
	}

	message = &pb.Message{}
	if err := proto.Unmarshal(buffer[:n], message); err != nil {
		return nil, err
	}
	return message, nil
}
