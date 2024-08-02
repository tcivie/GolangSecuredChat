package viewmodel

import (
	"client/internal/service"
	"fmt"
)

type AuthViewModel struct {
	loginService    *service.LoginService
	registerService *service.RegisterService
	Username        string
	onLogin         *func()
}

func NewAuthViewModel(commService *service.CommunicationService) *AuthViewModel {
	return &AuthViewModel{
		loginService:    service.NewLoginService(commService),
		registerService: service.NewRegisterService(commService),
	}
}

func (vm *AuthViewModel) SetOnLogin(callback func()) {
	vm.onLogin = &callback
}

func (vm *AuthViewModel) Login() (onLogin *func(), err error) {
	if vm.Username == "" {
		return nil, fmt.Errorf("username cannot be empty")
	}
	if err = vm.loginService.Login(vm.Username); err != nil {
		return nil, err
	}
	return vm.onLogin, nil
}

func (vm *AuthViewModel) Register() (onLogin *func(), err error) {
	if vm.Username == "" {
		return nil, fmt.Errorf("username cannot be empty")
	}
	if err := vm.registerService.Register(vm.Username); err != nil {
		return nil, err
	}
	return vm.Login()
}
