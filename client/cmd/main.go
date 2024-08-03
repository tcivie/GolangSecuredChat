package main

import (
	"client/internal/model"
	"client/internal/service"
	"client/internal/view"
	"client/internal/viewmodel"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/widget"
	"os"
	"time"
)

func main() {
	a := app.New()
	client, err := model.NewClient("localhost:8080", os.Getenv("PRIVATE_KEY"))
	if err == nil {
		commService := service.NewCommunicationService(client)

		loginVM := viewmodel.NewAuthViewModel(commService)
		loginView := view.NewLoginView(loginVM, a)

		userListVM := viewmodel.NewUserListViewModel(commService)
		userListView := view.NewUserListView(userListVM, a)

		chatVM := viewmodel.NewChatViewModel(commService)
		chatView := view.NewChatView(chatVM, a)

		loginVM.SetOnLogin(func() {
			userListView.Show()
			userListVM.FetchUsers()
			go func() {
				ticker := time.NewTicker(1 * time.Second)
				defer ticker.Stop()

				for range ticker.C {
					userListVM.FetchUsers()
					userListView.Update()
				}
			}()
			chatVM.WaitForHandshakeMessages()
		})

		userListVM.SetOnSelect(func(selectedUser string) {
			chatView.View()
			userListView.Hide()
			chatVM.SetCurrentChat(selectedUser)
			chatView.UpdateHeader(selectedUser)
			chatView.ReceiveMessages()
		})

		chatVM.SetOnBack(func() {
			chatView.Hide()
			userListView.Show()
		})

		loginView.Run()
		loginView.Show()
		a.Run()
	} else {
		a := app.New()
		errWindow := a.NewWindow("Error connecting to server")
		errorMessage := widget.NewLabel(err.Error())
		errWindow.SetContent(errorMessage)
		errWindow.ShowAndRun()
	}
}
