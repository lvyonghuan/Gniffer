package gniffer

import (
	"fmt"
	"sync/atomic"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

var OutPutChan = make(chan TreeRoot, 100)

func (n *NetCard) getData() {
	for {
		select {
		case <-n.stopCtx.Done():
			debugPrint("Stop analyzing on" + n.device.Description)
			return
		case <-n.reset:
			debugPrint("Reset analyzing on" + n.device.Description)
			n.bufferMu.Lock()
			n.buffer = n.buffer[:0]
			n.bufferMu.Unlock()
			atomic.StoreInt32(&n.nextID, 0)
		case packet := <-n.originDataChan:
			go n.analyzePacket(packet)
		}
	}
}

func (n *NetCard) analyzePacket(packet gopacket.Packet) {
	var treeRoot TreeRoot
	id := atomic.AddInt32(&n.nextID, 1)
	treeRoot.ID = int(id)

	defer func() {
		n.bufferMu.Lock()
		defer n.bufferMu.Unlock()
		n.buffer = append(n.buffer, treeRoot)
		OutPutChan <- treeRoot
		//debugPrint(treeRoot)
	}()

	// Ethernet layer
	leafEthernet, typ := getEthernetLayer(packet)
	if leafEthernet.Info != "" {
		treeRoot.Children = append(treeRoot.Children, leafEthernet)
	}

	// IP layer
	var protocol string
	switch typ {
	case "IPv4":
		leafIPv4, p := getIPv4Layer(packet)
		if leafIPv4.Info != "" {
			treeRoot.Children = append(treeRoot.Children, leafIPv4)
		}
		protocol = p
	case "IPv6":
		leafIPv6, p := getIPv6Layer(packet)
		if leafIPv6.Info != "" {
			treeRoot.Children = append(treeRoot.Children, leafIPv6)
		}
		protocol = p
	case "ARP":
		leafARP := getARP(packet)
		if leafARP.Info != "" {
			treeRoot.Children = append(treeRoot.Children, leafARP)
		}
		treeRoot.Protocol = "ARP"
		return
	case "LLC":
		leafLLC := getLLCLayer(packet)
		if leafLLC.Info != "" {
			treeRoot.Children = append(treeRoot.Children, leafLLC)
		}
		treeRoot.Protocol = "LLC"
		return
	default:
		debugPrint("Unknown type: " + typ)
		return
	}

	// Transport/Network layer
	switch protocol {
	case "TCP":
		leafTCP := getTCP(packet)
		if leafTCP.Info != "" {
			treeRoot.Children = append(treeRoot.Children, leafTCP)
		}
		treeRoot.Protocol = "TCP"
	case "UDP":
		leafUDP := getUDP(packet)
		if leafUDP.Info != "" {
			treeRoot.Children = append(treeRoot.Children, leafUDP)
		}
		treeRoot.Protocol = "UDP"
	case "ICMPv4":
		leafICMPv4 := getICMPv4Layer(packet)
		if leafICMPv4.Info != "" {
			treeRoot.Children = append(treeRoot.Children, leafICMPv4)
		}
		treeRoot.Protocol = "ICMPv4"
	case "ICMPv6":
		leafICMPv6 := getICMPv6Layer(packet)
		if leafICMPv6.Info != "" {
			treeRoot.Children = append(treeRoot.Children, leafICMPv6)
		}
		treeRoot.Protocol = "ICMPv6"
	default:
		debugPrint("Unknown protocol: " + protocol)
	}
}

func getEthernetLayer(packet gopacket.Packet) (Leaf, string) {
	etherNetLayer := packet.Layer(layers.LayerTypeEthernet)
	if etherNetLayer != nil {
		etherNet := etherNetLayer.(*layers.Ethernet)
		var leaf Leaf
		leaf.Info = "Source MAC: " + etherNet.SrcMAC.String() + "\n" + "Destination MAC: " + etherNet.DstMAC.String() + "\n" + "Ethernet type: " + etherNet.EthernetType.String()
		leaf.Hex = fmt.Sprintf("%x", etherNet.LayerContents())
		return leaf, etherNet.EthernetType.String()
	}
	return Leaf{}, ""
}

func getIPv4Layer(packet gopacket.Packet) (Leaf, string) {
	ipv4Layer := packet.Layer(layers.LayerTypeIPv4)
	if ipv4Layer != nil {
		ipv4 := ipv4Layer.(*layers.IPv4)
		var leaf Leaf
		leaf.Info = fmt.Sprintf("Version:%d\nIHL:%d\nTOS:%d\nLength:%d\nId:%d\nFlags:%d\nFragOffset:%d\nTTL:%d\nProtocol:%d\nChecksum:%d\nSource:%s\nDestination:%s", ipv4.Version, ipv4.IHL, ipv4.TOS, ipv4.Length, ipv4.Id, ipv4.Flags, ipv4.FragOffset, ipv4.TTL, ipv4.Protocol, ipv4.Checksum, ipv4.SrcIP.String(), ipv4.DstIP.String())
		leaf.Hex = fmt.Sprintf("%x", ipv4.LayerContents())
		return leaf, ipv4.Protocol.String()
	}
	return Leaf{}, ""
}

func getIPv6Layer(packet gopacket.Packet) (Leaf, string) {
	ipv6Layer := packet.Layer(layers.LayerTypeIPv6)
	if ipv6Layer != nil {
		ipv6 := ipv6Layer.(*layers.IPv6)
		var leaf Leaf
		leaf.Info = fmt.Sprintf("Version:%d\nLength:%d\nFlow Label:%d\nNext Header:%d\nHop Limit:%d\nSource Address:%s\nDestination Address:%s", ipv6.Version, ipv6.Length, ipv6.FlowLabel, ipv6.NextHeader, ipv6.HopLimit, ipv6.SrcIP.String(), ipv6.DstIP.String())
		leaf.Hex = fmt.Sprintf("%x", ipv6.LayerContents())
		return leaf, ipv6.NextHeader.String()
	}
	return Leaf{}, ""
}

func getARP(packet gopacket.Packet) Leaf {
	arpLayer := packet.Layer(layers.LayerTypeARP)
	if arpLayer != nil {
		arp := arpLayer.(*layers.ARP)
		var leaf Leaf
		leaf.Info = fmt.Sprintf("Operation:%d\nSourceHwAddress:%s\nSourceProtAddress:%s\nDstHwAddress:%s\nDstProtAddress:%s", arp.Operation, arp.SourceHwAddress, arp.SourceProtAddress, arp.DstHwAddress, arp.DstProtAddress)
		leaf.Hex = fmt.Sprintf("%x", arp.LayerContents())
		return leaf
	}
	return Leaf{}
}

func getLLCLayer(packet gopacket.Packet) Leaf {
	llcLayer := packet.Layer(layers.LayerTypeLLC)
	if llcLayer != nil {
		llc := llcLayer.(*layers.LLC)
		var leaf Leaf
		leaf.Info = fmt.Sprintf("DSAP:%d\nSSAP:%d\nControl:%d\nIG:%t\nCR:%t\n\nPayload:%v", llc.DSAP, llc.SSAP, llc.Control, llc.IG, llc.CR, llc.Payload)
		leaf.Hex = fmt.Sprintf("%x", llc.LayerContents())
		return leaf
	}
	return Leaf{}
}

func getICMPv4Layer(packet gopacket.Packet) Leaf {
	icmpv4Layer := packet.Layer(layers.LayerTypeICMPv4)
	if icmpv4Layer != nil {
		icmpv4 := icmpv4Layer.(*layers.ICMPv4)
		var leaf Leaf
		leaf.Info = "Type: " + icmpv4.TypeCode.String()
		leaf.Hex = fmt.Sprintf("%x", icmpv4.LayerContents())
		return leaf
	}
	return Leaf{}
}

func getICMPv6Layer(packet gopacket.Packet) Leaf {
	icmpv6Layer := packet.Layer(layers.LayerTypeICMPv6)
	if icmpv6Layer != nil {
		icmpv6 := icmpv6Layer.(*layers.ICMPv6)
		var leaf Leaf
		leaf.Info = fmt.Sprintf("Type:%d\nCode:%d\nChecksum:%d\nData:%s", icmpv6.TypeCode.Type(), icmpv6.TypeCode.Code(), icmpv6.Checksum, icmpv6.Payload)
		leaf.Hex = fmt.Sprintf("%x", icmpv6.LayerContents())
		return leaf
	}
	return Leaf{}
}

func getTCP(packet gopacket.Packet) Leaf {
	tcpLayer := packet.Layer(layers.LayerTypeTCP)
	if tcpLayer != nil {
		tcp := tcpLayer.(*layers.TCP)
		var leaf Leaf
		leaf.Info = fmt.Sprintf("From: %s\nTo: %s\nSeq: %d\nAck: %d\nDataOffset: %d\nFIN: %t\nSYN: %t\nRST: %t\nPSH: %t\nACK: %t\nURG: %t\nECE: %t\nCWR: %t\nNS: %t\nWindow: %d\nChecksum: %d\nUrgent: %d", tcp.SrcPort.String(), tcp.DstPort.String(), tcp.Seq, tcp.Ack, tcp.DataOffset, tcp.FIN, tcp.SYN, tcp.RST, tcp.PSH, tcp.ACK, tcp.URG, tcp.ECE, tcp.CWR, tcp.NS, tcp.Window, tcp.Checksum, tcp.Urgent)
		leaf.Hex = fmt.Sprintf("%x", tcp.LayerContents())
		return leaf
	}
	return Leaf{}
}

func getUDP(packet gopacket.Packet) Leaf {
	udpLayer := packet.Layer(layers.LayerTypeUDP)
	if udpLayer != nil {
		udp := udpLayer.(*layers.UDP)
		var leaf Leaf
		leaf.Info = fmt.Sprintf("From: %s\nTo: %s\nLength: %d\nChecksum: %d", udp.SrcPort.String(), udp.DstPort.String(), udp.Length, udp.Checksum)
		leaf.Hex = fmt.Sprintf("%x", udp.LayerContents())
		return leaf
	}
	return Leaf{}
}
