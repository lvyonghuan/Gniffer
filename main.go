package main

import (
	"gniffer/api"
	"gniffer/gniffer"
)

func main() {
	gniffer.GetNetCards()
	go gniffer.HandelData()
	api.InitRouter()
}
