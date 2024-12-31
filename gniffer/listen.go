package gniffer

import (
	"log"

	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
)

func (n *NetCard) listen() {
	snaplen := int32(65535)
	promisc := true
	timeout := pcap.BlockForever

	handle, err := pcap.OpenLive(n.Name, snaplen, promisc, timeout)
	if err != nil {
		log.Fatal(err)
	}
	defer handle.Close()

	debugPrint("Start listening on " + n.device.Description)
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())

	for {
		select {
		case <-n.stopCtx.Done():
			debugPrint("Stop listening on" + n.device.Description)
			return
		case data := <-packetSource.Packets():
			//debugPrint(data.String())
			n.originDataChan <- data
		}
	}
}
