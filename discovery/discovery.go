package discovery

import (
	"context"
	"fmt"
	"time"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
)

// Interval is how often we re-publish our mDNS records.
const Interval = time.Hour

// ServiceTag is used in our mDNS advertisements to discover other chat peers.
const ServiceTag = "pubsub-chat-frln"

type notifee struct {
	h host.Host
}

func (d *notifee) HandlePeerFound(pi peer.AddrInfo) {
	fmt.Printf("Discovered new peer: %s\n", pi.ID)
	err := d.h.Connect(context.Background(), pi)
	if err != nil {
		fmt.Printf("Error connecting to peer %s: %v\n", pi.ID, err)
		return
	}
}

func Setup(h host.Host) error {
	s := mdns.NewMdnsService(h, ServiceTag, &notifee{h: h})
	return s.Start()
}
