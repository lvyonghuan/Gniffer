package main

import (
	"context"
	"sync"

	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
)

type NetCard struct {
	Name string

	device pcap.Interface

	stopCtx        context.Context
	originDataChan chan gopacket.Packet
	reset          chan struct{}

	buffer   []data
	bufferMu sync.Mutex
	nextID   int32
}

type data struct {
	id int
	d  gopacket.Packet
}

type TreeRoot struct {
	ID         int    `json:"id"`
	Children   []Leaf `json:"children"`
	OriginData string `json:"originData"`
}

type Leaf struct {
	Info     any    `json:"info"`
	Hex      string `json:"hex"` //对应的16进制数据
	Children []Leaf `json:"children"`
}
