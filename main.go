package main

import (
	"gniffer/api"
	"gniffer/gniffer"
)

func main() {
	//analyze.RegisterOpenFlow()

	gniffer.GetNetCards()
	api.InitRouter()
}
