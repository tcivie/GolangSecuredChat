package view

import (
	"client/internal/model"
	"client/internal/viewmodel"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type ChatView struct {
	viewModel *viewmodel.ChatViewModel
	window    fyne.Window
	messages  *widget.List
	input     *widget.Entry
}

func NewChatView(vm *viewmodel.ChatViewModel) *ChatView {
	return &ChatView{viewModel: vm}
}

func (v *ChatView) Run() {
	a := app.New()
	v.window = a.NewWindow("Chat Application")

	v.messages = widget.NewList(
		func() int {
			return v.viewModel.GetMessageCount()
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("placeholder")
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			item.(*widget.Label).SetText(v.viewModel.GetMessageContent(id))
		},
	)

	v.input = widget.NewEntry()
	v.input.SetPlaceHolder("Type a message...")

	send := widget.NewButton("Send", func() {
		if v.input.Text != "" {
			v.viewModel.SendMessage(v.input.Text)
			v.input.SetText("")
			v.refreshMessageView()
		}
	})

	content := container.NewBorder(nil, container.NewBorder(nil, nil, nil, send, v.input), nil, nil, v.messages)
	v.window.SetContent(content)

	go v.receiveMessages()

	v.window.Resize(fyne.NewSize(400, 300))
	v.window.ShowAndRun()
}

func (v *ChatView) receiveMessages() {
	messageChan := make(chan model.Message)
	go v.viewModel.ReceiveMessages(messageChan)
	print("receiving messages")
	for message := range messageChan {
		print("adding message")
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
