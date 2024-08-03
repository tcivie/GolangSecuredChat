package viewmodel

import (
	"client/internal/service"
	"fmt"
	"os"
)

type AuthViewModel struct {
	loginService    *service.LoginService
	registerService *service.RegisterService
	commService     *service.CommunicationService
	Username        string
	onLogin         *func()
	//
	PrivateKeyPath string
	ServerAddress  string
}

func NewAuthViewModel(commService *service.CommunicationService) *AuthViewModel {
	return &AuthViewModel{
		loginService:    service.NewLoginService(commService),
		registerService: service.NewRegisterService(commService),
		commService:     commService,
		ServerAddress:   os.Getenv("SERVER_ADDRESS"),
	}
}

func (vm *AuthViewModel) SetOnLogin(callback func()) {
	vm.onLogin = &callback
}

func (vm *AuthViewModel) SetUsername(username string) {
	vm.Username = username
}

func (vm *AuthViewModel) SetServerAddress(serverAddress string) {
	vm.ServerAddress = serverAddress
}

func (vm *AuthViewModel) SetPrivateKeyPath(privateKeyPath string) error {
	vm.PrivateKeyPath = privateKeyPath
	return vm.commService.SetPrivateKeyPath(privateKeyPath)
}

func (vm *AuthViewModel) Login() (onLogin *func(), err error) {
	if !vm.commService.GetClient().IsConnected() {
		if err := vm.connectToServer(); err != nil {
			return nil, nil
		}
	}

	if vm.Username == "" {
		return nil, fmt.Errorf("username cannot be empty")
	}
	if err = vm.loginService.Login(vm.Username); err != nil {
		return nil, err
	}
	return vm.onLogin, nil
}

func (vm *AuthViewModel) Register() (err error) {
	if !vm.commService.GetClient().IsConnected() {
		if err := vm.connectToServer(); err != nil {
			return nil
		}
	}

	if vm.Username == "" {
		return fmt.Errorf("username cannot be empty")
	}
	if err := vm.registerService.Register(vm.Username); err != nil {
		return err
	}
	vm.commService.GetClient().Username = vm.Username
	if err = vm.commService.SetPrivateKeyPath(vm.PrivateKeyPath); err != nil {
		return err
	}
	return
}

func (vm *AuthViewModel) connectToServer() error {
	if err := vm.commService.GetClient().MakeConnection(vm.ServerAddress, vm.PrivateKeyPath); err != nil {
		return err
	}
	if vm.commService.GetClient().IsConnected() {
		vm.commService.StartHandlingMessages()
	}
	return nil
}
