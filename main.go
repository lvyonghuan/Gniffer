package main

import (
	"gniffer/api"
	"gniffer/gniffer"
)

func main() {
	gniffer.GetNetCards()
	api.InitRouter()
}
