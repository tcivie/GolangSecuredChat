package actions

import (
	pb "server/resources/proto"
)

type MessageContext struct {
	strategy MessageHandler
}

func NewMessageContext(strategy MessageHandler) *MessageContext {
	return &MessageContext{
		strategy: strategy,
	}
}

func (c *MessageContext) ExecuteStrategy(message *pb.Message) error {
	return c.strategy.handleMessage(message)
}
