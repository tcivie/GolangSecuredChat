package main

import (
	"client/internal/model"
	"client/internal/view"
	"client/internal/viewmodel"
	"log"
)

func main() {
	client, err := model.NewClient("localhost:8080")
	if err != nil {
		log.Fatal(err)
	}

	chatVM := viewmodel.NewChatViewModel(client)
	chatView := view.NewChatView(chatVM)

	chatView.Run()
}
