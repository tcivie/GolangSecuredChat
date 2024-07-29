package view

import (
	"client/internal/viewmodel"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type UserListView struct {
	viewModel *viewmodel.UserListViewModel
	app       fyne.App
	window    fyne.Window
	userList  *widget.List
}

func NewUserListView(vm *viewmodel.UserListViewModel, app fyne.App) *UserListView {
	return &UserListView{
		viewModel: vm,
		app:       app,
	}
}

func (v *UserListView) Run() {
	v.window = v.app.NewWindow("Active Users")

	v.userList = widget.NewList(
		func() int { return len(v.viewModel.Users) },
		func() fyne.CanvasObject { return widget.NewLabel("User") },
		func(id widget.ListItemID, item fyne.CanvasObject) {
			item.(*widget.Label).SetText(v.viewModel.Users[id])
		},
	)

	v.userList.OnSelected = func(id widget.ListItemID) {
		selectedUser := v.viewModel.Users[id]
		v.onUserSelected(selectedUser)
		v.userList.Unselect(id)
	}

	content := container.NewBorder(nil, nil, nil, nil, v.userList)
	v.window.SetContent(content)
	v.window.Resize(fyne.NewSize(300, 400))
}

func (v *UserListView) Update() {
	v.userList.Refresh()
}

func (v *UserListView) GetUsers() {
	v.viewModel.FetchUsers()
}

func (v *UserListView) Show() {
	v.Run()
	v.window.Show()
}

func (v *UserListView) Hide() {
	//v.window.Hide()
	v.window.Close()
}

func (v *UserListView) onUserSelected(selectedUser string) {
	onSelect, err := v.viewModel.SelectedUser(selectedUser)
	if err != nil {
		// Handle error (e.g., log it)
		return
	}
	if onSelect != nil {
		(*onSelect)(selectedUser)
	}
}
