package main

import (
	"client/internal/service"
	"client/internal/view"
	"client/internal/viewmodel"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/widget"
	"time"
)

func main() {
	a := app.New()
	chatService, err := service.NewChatService("localhost:8080", "client/resources/private2")
	if err == nil {
		loginVM := viewmodel.NewAuthViewModel(chatService.Client)
		loginView := view.NewLoginView(loginVM, a)

		userListVM := viewmodel.NewUserListViewModel(chatService)
		userListView := view.NewUserListView(userListVM, a)

		chatVM := viewmodel.NewChatViewModel(chatService)
		chatView := view.NewChatView(chatVM, a)

		loginVM.SetOnLogin(func() {
			userListView.Show()
			userListVM.FetchUsers()

			go func() {
				ticker := time.NewTicker(10 * time.Second)
				defer ticker.Stop()

				for range ticker.C {
					userListView.GetUsers()
					userListView.Update()
				}
			}()
		})

		userListVM.SetOnSelect(func(selectedUser string) {
			userListView.Hide()
			chatView.View()
			chatVM.SetCurrentChat(selectedUser)
			go chatView.ReceiveMessages()
			chatView.UpdateHeader(selectedUser)
		})

		chatVM.SetOnBack(func() {
			chatView.Hide()
			userListView.Show()
		})
		//chatView.Run()
		//userListView.Run()
		loginView.Run()
		loginView.Show()
		a.Run()
	} else {
		a := app.New()
		errWindow := a.NewWindow("Error connecting to server")
		errorMessage := widget.NewLabel(err.Error())
		errWindow.SetContent(errorMessage)
		errWindow.ShowAndRun()
		a.Run()
	}
}
