package gniffer

import (
	"context"
	"sync"

	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
)

var cards = make(map[string]NetCard)

// GetNetCards Get all network cards
func GetNetCards() []NetCard {
	var netCards []NetCard

	devices, err := pcap.FindAllDevs()
	if err != nil {
		panic(err)
	}

	for _, device := range devices {
		//debugPrint(device.Description)
		var NetCard = NetCard{
			Description: device.Description,
			Name:        device.Name,
			device:      device,
		}

		netCards = append(netCards, NetCard)
		cards[device.Name] = NetCard
	}

	return netCards
}

func (n *NetCard) Init() {
	n.stopCtx = context.Background()
	n.originDataChan = make(chan gopacket.Packet, 100)
	n.reset = make(chan struct{})
	n.bufferMu = sync.Mutex{}
}
