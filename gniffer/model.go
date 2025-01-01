package gniffer

import (
	"context"
	"sync"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
)

type NetCard struct {
	Description string `json:"description"`
	Name        string `json:"name"`

	device pcap.Interface

	stopCtx        context.Context
	stopCancelFunc context.CancelFunc
	originDataChan chan gopacket.Packet
	reset          chan struct{}

	buffer   []TreeRoot
	bufferMu sync.Mutex
	nextID   int32
}

type TreeRoot struct {
	ID         int    `json:"id"`
	Children   []Leaf `json:"children"`
	OriginData string `json:"originData"`

	Time        string `json:"time"`
	time        time.Time
	Source      string `json:"source"`
	Destination string `json:"destination"`
	Protocol    string `json:"protocol"`
	Length      int    `json:"length"`
	Info        string `json:"info"`
}

type Leaf struct {
	Name string `json:"name"`
	Info any    `json:"info"`
	Hex  string `json:"hex"` //对应的16进制数据
}
