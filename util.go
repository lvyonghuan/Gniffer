package main

import (
	"log"
)

const isDebug = true

func debugPrint(info any) {
	if isDebug {
		log.Println("\u001B[32m", info, "\u001B[0m")
	}
}
