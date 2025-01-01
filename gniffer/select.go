package gniffer

import (
	"context"
	"errors"
	"sync"

	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
)

var cards = make(map[string]*NetCard)

// GetNetCards Get all network cards
func GetNetCards() []NetCard {
	var netCards []NetCard

	devices, err := pcap.FindAllDevs()
	if err != nil {
		panic(err)
	}

	for _, device := range devices {
		debugPrint("Name: " + device.Name + " Description: " + device.Description)
		var NetCard = NetCard{
			Description: device.Description,
			Name:        device.Name,
			device:      device,
		}

		netCards = append(netCards, NetCard)
		cards[device.Name] = &NetCard
	}

	return netCards
}

func GetCards() []NetCard {
	var netCards []NetCard
	for _, card := range cards {
		netCards = append(netCards, *card)
	}
	return netCards
}

func (n *NetCard) Init() {
	debugPrint("Init card" + n.device.Description)
	n.stopCtx, n.stopCancelFunc = context.WithCancel(context.Background())
	n.originDataChan = make(chan gopacket.Packet, 100)
	n.reset = make(chan struct{})
	n.bufferMu = sync.Mutex{}
	n.nextID = 0
}

func Start(cardName string) error {
	card, isExist := cards[cardName]
	if !isExist {
		return errors.New("card not found")
	}

	card.Init()
	go card.listen()
	go card.getData()

	return nil
}

func Stop(cardName string) {
	card := cards[cardName]
	card.stopCancelFunc()

	*beforeHandelBuffer = (*beforeHandelBuffer)[:0]
	frontBuffer = frontBuffer[:0]
}
