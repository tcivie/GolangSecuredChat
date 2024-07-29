package view

import (
	"client/internal/viewmodel"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type AuthView struct {
	viewModel  *viewmodel.AuthViewModel
	app        fyne.App
	window     fyne.Window
	username   *widget.Entry
	status     *widget.Label
	statusChan chan string
}

func NewLoginView(vm *viewmodel.AuthViewModel, app fyne.App) *AuthView {
	return &AuthView{
		viewModel:  vm,
		app:        app,
		statusChan: make(chan string),
	}
}

func (v *AuthView) Run() {
	v.window = v.app.NewWindow("CryptoChat - Secure Messaging")

	// Create logo
	logo := canvas.NewImageFromResource(theme.FyneLogo())
	logo.SetMinSize(fyne.NewSize(100, 100))

	// App name and tagline
	appName := widget.NewLabel("CryptoChat")
	appName.TextStyle = fyne.TextStyle{Bold: true}
	tagline := widget.NewLabel("Secure, Simple, Swift")

	// Information about the app
	infoTitle := widget.NewRichTextFromMarkdown("CryptoChat uses advanced encryption to ensure your privacy:")
	infoTitle.Segments[0].(*widget.TextSegment).Style.Alignment = fyne.TextAlignCenter

	infoList := widget.NewRichText(
		&widget.TextSegment{
			Text: "• Username-only login with stored private keys\n" +
				"• TLS-secured server connection\n" +
				"• End-to-end encryption for all messages\n" +
				"• Public key validation for user authentication",
			Style: widget.RichTextStyle{Alignment: fyne.TextAlignLeading},
		},
	)

	info := container.NewVBox(infoTitle, infoList)

	v.username = widget.NewEntry()
	v.username.SetPlaceHolder("Enter username")

	loginButton := widget.NewButton("Secure Login", func() {
		go v.attemptLogin()
	})
	loginButton.Importance = widget.HighImportance

	registerButton := widget.NewButton("Register", func() {
		go v.attemptRegister()
	})

	v.status = widget.NewLabel("")

	content := container.NewVBox(
		container.NewCenter(logo),
		container.NewCenter(appName),
		container.NewCenter(tagline),
		layout.NewSpacer(),
		widget.NewCard("", "", info),
		layout.NewSpacer(),
		v.username,
		loginButton,
		registerButton,
		v.status,
	)

	// Wrap the content in a padded container
	paddedContent := container.NewPadded(content)

	v.window.SetContent(paddedContent)
	v.window.Resize(fyne.NewSize(400, 600))

	go v.listenForStatusUpdates()
}

func (v *AuthView) attemptLogin() {
	username := v.username.Text
	v.viewModel.Username = username
	onLogin, err := v.viewModel.Login()
	v.updateStatus(err, "Login", onLogin)
}

func (v *AuthView) attemptRegister() {
	username := v.username.Text
	v.viewModel.Username = username
	onLogin, err := v.viewModel.Register()
	v.updateStatus(err, "Registration", onLogin)
}

func (v *AuthView) updateStatus(err error, action string, onLogin *func()) {
	if err != nil {
		v.statusChan <- action + " failed: " + err.Error()
	} else {
		v.statusChan <- action + " successful!"
		v.window.Close()
		if onLogin != nil {
			(*onLogin)()
		}
	}
}

func (v *AuthView) listenForStatusUpdates() {
	for status := range v.statusChan {
		v.status.SetText(status)
	}
}

func (v *AuthView) Show() {
	v.window.Show()
}
