package main

import (
	"client/internal/model"
	"client/internal/service"
	"client/internal/view"
	"client/internal/viewmodel"
	"fyne.io/fyne/v2/app"
	"time"
)

func main() {
	a := app.New()
	client := model.NewClient()
	if true {
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
	}
}
