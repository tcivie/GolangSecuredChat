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

type ChatView struct {
	viewModel *viewmodel.ChatViewModel
	app       fyne.App
	window    fyne.Window
	messages  *widget.List
	input     *widget.Entry
	header    *widget.Label
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

	// Header with current chat partner
	v.header = widget.NewLabel("Chat with: ")
	v.header.TextStyle = fyne.TextStyle{Bold: true}

	// Header
	header := container.NewVBox(
		container.NewHBox(
			logo, layout.NewSpacer(),
			appName, layout.NewSpacer(),
			widget.NewButtonWithIcon("", theme.NavigateBackIcon(), v.navigateBack),
		),
		v.header,
	)

	v.messages = widget.NewList(
		func() int {
			return v.viewModel.GetMessageCount()
		},
		func() fyne.CanvasObject {
			return container.NewHBox(
				widget.NewIcon(theme.AccountIcon()),
				widget.NewLabel("Message placeholder"),
			)
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			message := v.viewModel.GetMessageContent(id)
			contentLabel := item.(*fyne.Container).Objects[1].(*widget.Label)
			contentLabel.SetText(message)
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

	v.window.SetOnClosed(func() {
		v.viewModel.StopReceivingMessages()
	})

	v.window.Resize(fyne.NewSize(400, 600))
}

func (v *ChatView) View() {
	v.Run()
	v.window.Show()
}

func (v *ChatView) ReceiveMessages() {
	go v.handleIncomingMessages()
	v.viewModel.StartReceivingMessages()
}

func (v *ChatView) handleIncomingMessages() {
	for message := range v.viewModel.GetMessageChan() {
		if message.Content != "" {
			v.viewModel.AddMessage(message)
			v.refreshMessageView()
		}
	}
}

func (v *ChatView) Hide() {
	v.viewModel.StopReceivingMessages()
	v.window.Close()
}

func (v *ChatView) submitContent(content string) {
	if content != "" {
		go v.viewModel.SendMessage(content)
		v.input.SetText("")
		v.refreshMessageView()
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

func (v *ChatView) UpdateHeader(username string) {
	v.header.SetText("Chat with: " + username)
}

func (v *ChatView) navigateBack() {
	v.viewModel.Back()
}
