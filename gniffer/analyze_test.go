package gniffer

import (
	"fmt"
	"testing"
	"time"
)

func TestAnalyze(t *testing.T) {
	t.Log("Test case")
	cards := getNetCards()
	if len(cards) == 0 {
		t.Error("No network cards found")
	}

	card := cards[5]

	card.init()
	go card.listen()
	go card.getData()

	for {
		if len(card.buffer) >= 1000 {
			card.stopCtx.Done()
			time.Sleep(1 * time.Second)

			for _, tree := range card.buffer {
				fmt.Println(tree)
			}

			t.Log("Test case end")
			return
		}
	}
}
