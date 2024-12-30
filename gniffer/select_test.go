package gniffer

import "testing"

func TestGetNetCards(t *testing.T) {
	cards := GetNetCards()
	if len(cards) == 0 {
		t.Error("No network cards found")
	}

	for _, card := range cards {
		if card.Name == "" {
			t.Error("Invalid network card name")
		}
	}
	t.Log("Network cards found")
}
