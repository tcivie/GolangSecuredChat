package view

import (
	"client/internal/model"
	"client/internal/viewmodel"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type ChatView struct {
	viewModel *viewmodel.ChatViewModel
	app       fyne.App
	window    fyne.Window
	messages  *widget.List
	input     *widget.Entry
}

func NewChatView(vm *viewmodel.ChatViewModel, app fyne.App) *ChatView {
	return &ChatView{viewModel: vm, app: app}
}

func (v *ChatView) Run() {
	v.window = v.app.NewWindow("CryptoChat - Secure Messaging")

	// Create logo
	logo := canvas.NewImageFromResource(theme.FyneLogo())
	logo.SetMinSize(fyne.NewSize(40, 40))

	// App name
	appName := widget.NewLabel("CryptoChat")
	appName.TextStyle = fyne.TextStyle{Bold: true}

	// Header
	header := container.NewHBox(logo, layout.NewSpacer(), appName, layout.NewSpacer())

	v.messages = widget.NewList(
		func() int {
			return v.viewModel.GetMessageCount()
		},
		func() fyne.CanvasObject {
			return container.NewHBox(
				widget.NewIcon(theme.AccountIcon()),
				widget.NewLabel("User:"),
				widget.NewLabel("Message placeholder"),
			)
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			message := v.viewModel.GetMessageContent(id)
			user, content := parseMessage(message)
			userLabel := item.(*fyne.Container).Objects[1].(*widget.Label)
			contentLabel := item.(*fyne.Container).Objects[2].(*widget.Label)

			userLabel.SetText(user + ":")
			userLabel.TextStyle = fyne.TextStyle{Bold: true}

			contentLabel.SetText(content)
		},
	)

	v.input = widget.NewEntry()
	v.input.SetPlaceHolder("Type a message...")

	send := widget.NewButtonWithIcon("Send", theme.MailSendIcon(), func() {
		v.submitContent(v.input.Text)
	})
	send.Importance = widget.HighImportance

	inputContainer := container.NewBorder(nil, nil, nil, send, v.input)

	content := container.NewBorder(header, inputContainer, nil, nil, container.NewPadded(v.messages))

	// Add some padding around the entire content
	paddedContent := container.NewPadded(content)

	v.window.SetContent(paddedContent)

	v.window.Resize(fyne.NewSize(400, 600))
}

func (v *ChatView) View() {
	v.window.Show()
}

func (v *ChatView) submitContent(content string) {
	if content != "" {
		v.viewModel.SendMessage(content)
		v.input.SetText("")
		v.refreshMessageView()
	}
}

func (v *ChatView) ReceiveMessages() {
	messageChan := make(chan model.Message)
	go v.viewModel.ReceiveMessages(messageChan)
	for message := range messageChan {
		if message.Content != "" {
			v.refreshMessageView()
		}
	}
}

func (v *ChatView) refreshMessageView() {
	v.messages.Refresh()
	v.window.Canvas().Content().Refresh()
	v.scrollToBottom()
}

func (v *ChatView) scrollToBottom() {
	if v.messages.Length() > 0 {
		v.messages.ScrollTo(v.messages.Length() - 1)
	}
}

func parseMessage(message string) (string, string) {
	// Implement message parsing logic here
	// For now, we'll just return placeholder values
	return "You", message
}
