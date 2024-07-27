package actions

import pb "server/resources/proto"

type MessageHandler interface {
	handleMessage(message *pb.Message) error
}
