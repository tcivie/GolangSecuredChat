package view

import (
	"client/internal/viewmodel"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type UserListView struct {
	viewModel *viewmodel.UserListViewModel
	app       fyne.App
	window    fyne.Window
	userList  *widget.List
	username  *widget.Label
	status    *widget.Label
}

func NewUserListView(vm *viewmodel.UserListViewModel, app fyne.App) *UserListView {
	return &UserListView{
		viewModel: vm,
		app:       app,
	}
}

func (v *UserListView) Run() {
	v.window = v.app.NewWindow("CryptoChat - Active Users")

	// Header
	v.username = widget.NewLabel(fmt.Sprintf("Logged in as: %s", v.viewModel.GetCurrentUsername()))
	v.username.TextStyle = fyne.TextStyle{Bold: true}

	refreshButton := widget.NewButtonWithIcon("", theme.ViewRefreshIcon(), func() {
		v.GetUsers()
	})

	header := container.NewBorder(nil, nil, nil, refreshButton,
		container.NewVBox(
			widget.NewLabel("CryptoChat"),
			v.username,
		),
	)

	// User list
	v.userList = widget.NewList(
		func() int { return len(v.viewModel.Users) },
		func() fyne.CanvasObject {
			return container.NewHBox(
				widget.NewIcon(theme.AccountIcon()),
				widget.NewLabel("User"),
			)
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			item.(*fyne.Container).Objects[1].(*widget.Label).SetText(v.viewModel.Users[id])
		},
	)

	v.userList.OnSelected = func(id widget.ListItemID) {
		selectedUser := v.viewModel.Users[id]
		v.onUserSelected(selectedUser)
		v.userList.Unselect(id)
	}

	// Status bar
	v.status = widget.NewLabel("Ready")

	// Layout
	content := container.NewBorder(header, v.status, nil, nil, v.userList)
	v.window.SetContent(content)
	v.window.Resize(fyne.NewSize(300, 400))
}

func (v *UserListView) Update() {
	v.userList.Refresh()
	v.username.SetText(fmt.Sprintf("Logged in as: %s", v.viewModel.GetCurrentUsername()))
	v.status.SetText(fmt.Sprintf("%d users online", len(v.viewModel.Users)))
}

func (v *UserListView) GetUsers() {
	v.status.SetText("Fetching users...")
	v.viewModel.FetchUsers()
	v.Update()
}

func (v *UserListView) Show() {
	v.Run()
	v.GetUsers() // Fetch users when showing the view
	v.window.Show()
}

func (v *UserListView) Hide() {
	v.window.Close()
}

func (v *UserListView) onUserSelected(selectedUser string) {
	v.status.SetText(fmt.Sprintf("Connecting to %s...", selectedUser))
	onSelect, err := v.viewModel.SelectedUser(selectedUser)
	if err != nil {
		v.status.SetText(fmt.Sprintf("Error: %s", err.Error()))
		return
	}
	if onSelect != nil {
		(*onSelect)(selectedUser)
	}
}
