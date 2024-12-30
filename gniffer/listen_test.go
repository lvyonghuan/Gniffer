package gniffer

import (
	"testing"
	"time"

	"github.com/google/gopacket"
)

func TestListen(t *testing.T) {
	t.Log("Test case")
	cards := GetNetCards()
	if len(cards) == 0 {
		t.Error("No network cards found")
	}

	card := cards[5]

	card.Init()
	go card.listen()

	var buffer []gopacket.Packet
	for {
		select {
		case data := <-card.originDataChan:
			buffer = append(buffer, data)
			t.Log(data.String())
		default:
			if len(buffer) >= 1 {
				card.stopCtx.Done()
				time.Sleep(1 * time.Second)
				t.Log("Test case end")
				return
			}
		}
	}
}
