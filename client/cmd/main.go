package main

import (
	"client/internal/service"
	"client/internal/view"
	"client/internal/viewmodel"
	"log"
)

func main() {
	chatService, err := service.NewChatService("localhost:8080")
	if err != nil {
		log.Fatal(err)
	}

	chatVM := viewmodel.NewChatViewModel(chatService)
	chatView := view.NewChatView(chatVM)

	chatView.Run()
}
