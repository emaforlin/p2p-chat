package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/libp2p/go-libp2p"
	pubsub "github.com/libp2p/go-libp2p-pubsub"

	"github.com/emaforlin/p2p-chatroom/chatroom"
	"github.com/emaforlin/p2p-chatroom/discovery"
	"github.com/emaforlin/p2p-chatroom/ui"
)

func main() {
	roomFlag := flag.String("room", "TheWest", "Name of the chat room to join")
	nickFlag := flag.String("nick", "McQueen96", "Nickname to use in the chat room")
	flag.Parse()

	ctx := context.Background()

	h, err := libp2p.New(libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/0"))
	if err != nil {
		panic(err)
	}

	ps, err := pubsub.NewGossipSub(ctx, h)
	if err != nil {
		panic(err)
	}

	if err := discovery.Setup(h); err != nil {
		panic(err)
	}

	cr, err := chatroom.Join(ctx, ps, *roomFlag, h.ID(), *nickFlag)
	if err != nil {
		panic(err)
	}

	chatUI := ui.NewChatUI(cr)
	if err := chatUI.Run(); err != nil {
		fmt.Printf("Error running TUI: %v\n", err)
		return
	}
}
