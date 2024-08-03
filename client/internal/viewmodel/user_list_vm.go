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
	//
	commService *service.CommunicationService
}

func NewUserListViewModel(commService *service.CommunicationService) *UserListViewModel {
	return &UserListViewModel{
		chatService: service.NewChatService(commService),
		Users:       []string{},
		commService: commService,
	}
}

func (vm *UserListViewModel) FetchUsers() {
	users, err := vm.chatService.GetUserList()
	if err != nil {
		// Handle error (e.g., log it)
		fmt.Printf("Error fetching users: %v\n", err)
		return
	}
	// filter out the current user
	currentUsername := vm.commService.GetUsername()
	for i, user := range users {
		if user == currentUsername {
			users = append(users[:i], users[i+1:]...)
			break
		}
	}
	vm.Users = users
}

func (vm *UserListViewModel) SetOnSelect(callback func(string)) {
	vm.onSelect = &callback
}

func (vm *UserListViewModel) SelectedUser(user string) (onLogin *func(string), err error) {
	if user == "" {
		return nil, fmt.Errorf("username cannot be empty")
	}
	// Handle selected user
	return vm.onSelect, nil
}
