package main

import (
	"client/internal/service"
	"client/internal/view"
	"client/internal/viewmodel"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/widget"
)

func main() {
	a := app.New()
	chatService, err := service.NewChatService("localhost:8080", "client/resources/private")
	if err == nil {
		loginVM := viewmodel.NewAuthViewModel(chatService.Client)
		loginView := view.NewLoginView(loginVM, a)

		chatVM := viewmodel.NewChatViewModel(chatService)
		chatView := view.NewChatView(chatVM, a)

		loginVM.SetOnLogin(func() {
			chatView.View()
		})
		chatView.Run()
		loginView.Run()
	} else {
		errWindow := a.NewWindow("Error connecting to server")
		errorMessage := widget.NewLabel(err.Error())
		errWindow.SetContent(errorMessage)
		errWindow.ShowAndRun()
	}
	a.Run()
}
