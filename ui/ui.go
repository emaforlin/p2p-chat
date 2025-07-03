package ui

import (
	"fmt"
	"io"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/emaforlin/p2p-chatroom/chatroom"
)

type ChatUI struct {
	cr        *chatroom.Room
	peersList *tview.TextView

	msgW    io.Writer
	app     *tview.Application
	inputCh chan string
	doneCh  chan struct{}
}

func NewChatUI(cr *chatroom.Room) *ChatUI {
	app := tview.NewApplication()

	msgBox := tview.NewTextView()
	msgBox.
		SetDynamicColors(true).
		SetBorder(true).
		SetTitle(fmt.Sprintf("Room: %s", cr.RoomName()))

	msgBox.SetChangedFunc(func() {
		app.Draw()
	})

	inputCh := make(chan string, 32)
	input := tview.NewInputField().
		SetLabel(fmt.Sprintf("<%s> says: ", cr.SelfNick())).
		SetFieldWidth(0).
		SetFieldBackgroundColor(tcell.ColorBlack)

	input.SetDoneFunc(func(key tcell.Key) {
		if key != tcell.KeyEnter {
			return
		}
		line := input.GetText()
		if len(line) == 0 {
			return
		}

		if line == "/quit" {
			app.Stop()
			return
		}

		inputCh <- line
		input.SetText("")
	})

	peerList := tview.NewTextView()
	peerList.SetBorder(true)
	peerList.SetTitle("Connected Peers")
	peerList.SetChangedFunc(func() {
		app.Draw()
	})

	chatPanel := tview.NewFlex().
		AddItem(msgBox, 0, 1, false).
		AddItem(peerList, 20, 1, false)

	flex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(chatPanel, 0, 1, false).
		AddItem(input, 1, 1, true)

	app.SetRoot(flex, true)

	return &ChatUI{
		cr:        cr,
		msgW:      msgBox,
		peersList: peerList,
		app:       app,
		inputCh:   inputCh,
		doneCh:    make(chan struct{}),
	}
}

func (ui *ChatUI) Run() error {
	go ui.handleEvents()
	return ui.app.Run()
}

func (ui *ChatUI) displaySelfMessage(msg string) {
	prompt := withColor("yellow", fmt.Sprintf("<%s>", ui.cr.SelfNick()))
	fmt.Fprintf(ui.msgW, "%s %s\n", prompt, msg)
}

func (ui *ChatUI) displayMessage(msg *chatroom.Message) {
	prompt := withColor("green", fmt.Sprintf("<%s>", msg.SenderNick))
	fmt.Fprintf(ui.msgW, "%s %s\n", prompt, msg.Message)
}

func (ui *ChatUI) refreshPeers() {
	peers := ui.cr.ListPeers()

	ui.peersList.Clear()

	for _, p := range peers {
		fmt.Fprintf(ui.peersList, "%s\n", shortID(p.String()))
	}

	ui.app.Draw()
}

func (ui *ChatUI) handleEvents() {
	peerRefreshTicker := time.NewTicker(time.Second)
	defer peerRefreshTicker.Stop()

	for {
		select {
		case msg := <-ui.inputCh:
			// Publish the message to the chatroom
			if err := ui.cr.Publish(msg); err != nil {
				fmt.Fprintf(ui.msgW, "Error sending message: %v\n", err)
				continue
			}
			ui.displaySelfMessage(msg)

		case msg := <-ui.cr.Messages:
			// Display messages from other peers
			ui.displayMessage(msg)

		case <-peerRefreshTicker.C:
			ui.refreshPeers()

		case <-ui.cr.Context().Done():
			return

		case <-ui.doneCh:
			return
		}
	}
}

func withColor(color, msg string) string {
	return fmt.Sprintf("[%s]%s[-]", color, msg)
}

func shortID(id string) string {
	if len(id) > 8 {
		return id[:8]
	}
	return id
}
