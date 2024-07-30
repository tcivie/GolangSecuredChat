package viewmodel

import (
	"client/internal/model"
	"client/internal/service"
	"fmt"
)

type UserListViewModel struct {
	chatService *service.ChatService
	Users       []string
	onSelect    *func(string)
	chatters    *map[string]model.Chatter
}

func NewUserListViewModel(service *service.ChatService) *UserListViewModel {
	return &UserListViewModel{
		chatService: service,
		Users:       []string{},
	}
}

func (vm *UserListViewModel) WaitForMessages() {
	go func() {
		for {
			message, err := vm.chatService.ReceiveMessage()
			if err != nil {
				continue
			}

		}
	}()

}

func (vm *UserListViewModel) FetchUsers() {
	users, err := vm.chatService.GetUserList()
	if err != nil {
		// Handle error (e.g., log it)
		return
	}
	vm.Users = users
}

func (vm *UserListViewModel) SetOnSelect(callback func(string)) {
	vm.onSelect = &callback
}

func (vm *UserListViewModel) SelectedUser(user string) (onLogin *func(string), err error) {
	if user == "" {
		// Handle error (e.g., log it)
		return nil, fmt.Errorf("username cannot be empty")
	}
	// Handle selected user
	return vm.onSelect, nil
}
