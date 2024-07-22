package viewmodel

import (
	"client/internal/model"
	"client/internal/service"
	"fmt"
)

type LoginViewModel struct {
	loginService *service.LoginService
	Username     string
	onLogin      *func()
}

func NewLoginViewModel(client *model.Client) *LoginViewModel {
	return &LoginViewModel{
		loginService: service.NewLoginService(client),
	}
}

func (vm *LoginViewModel) SetOnLogin(callback func()) {
	vm.onLogin = &callback
}

func (vm *LoginViewModel) Login() (onLogin *func(), err error) {
	if vm.Username == "" {
		return nil, fmt.Errorf("username cannot be empty")
	}
	err = vm.loginService.Login()
	return vm.onLogin, err
}
