package main

import (
	"context"
	"sync"

	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
)

// Get all network cards
func getNetCards() []NetCard {
	var netCards []NetCard

	devices, err := pcap.FindAllDevs()
	if err != nil {
		panic(err)
	}

	for _, device := range devices {
		debugPrint(device.Description)
		netCards = append(netCards, NetCard{Name: device.Name, device: device})
	}

	return netCards
}

func (n *NetCard) init() {
	n.stopCtx = context.Background()
	n.originDataChan = make(chan gopacket.Packet, 100)
	n.reset = make(chan struct{})
	n.bufferMu = sync.Mutex{}
}
