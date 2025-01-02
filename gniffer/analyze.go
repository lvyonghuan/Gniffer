package gniffer

import (
	"encoding/binary"
	"fmt"
	"sync/atomic"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

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
	treeRoot.Time = packet.Metadata().Timestamp.String()
	treeRoot.Length = packet.Metadata().Length

	defer func() {
		n.bufferMu.Lock()
		defer n.bufferMu.Unlock()
		n.buffer = append(n.buffer, treeRoot)
		//debugPrint(treeRoot)
	}()

	// Ethernet layer
	leafEthernet, typ, src, dst := getEthernetLayer(packet)
	if leafEthernet.Info != nil {
		treeRoot.Children = append(treeRoot.Children, leafEthernet)
		treeRoot.Source = src
		treeRoot.Destination = dst
	}

	// IP layer
	var protocol string
	switch typ {
	case IPv4:
		leafIPv4, p, src, dst := getIPv4Layer(packet)
		if leafIPv4.Info != nil {
			treeRoot.Children = append(treeRoot.Children, leafIPv4)
			treeRoot.Source = src
			treeRoot.Destination = dst
		}
		protocol = p
	case IPv6:
		leafIPv6, p, src, dst := getIPv6Layer(packet)
		if leafIPv6.Info != nil {
			treeRoot.Children = append(treeRoot.Children, leafIPv6)
			treeRoot.Source = src
			treeRoot.Destination = dst
		}
		protocol = p
	case ARP:
		leafARP := getARP(packet)
		if leafARP.Info != nil {
			treeRoot.Children = append(treeRoot.Children, leafARP)
		}
		treeRoot.Protocol = ARP
		return
	case LLC:
		leafLLC := getLLCLayer(packet)
		if leafLLC.Info != nil {
			treeRoot.Children = append(treeRoot.Children, leafLLC)
		}
		treeRoot.Protocol = LLC
		return
	default:
		debugPrint("Unknown type: " + typ)
		return
	}

	// Transport/Network layer
	var tcpContent []byte
	switch protocol {
	case TCP:
		leafTCP, content := getTCP(packet)
		if leafTCP.Info != nil {
			treeRoot.Children = append(treeRoot.Children, leafTCP)
		}
		treeRoot.Protocol = TCP
		tcpContent = content
	case UDP:
		leafUDP := getUDP(packet)
		if leafUDP.Info != nil {
			treeRoot.Children = append(treeRoot.Children, leafUDP)
		}
		treeRoot.Protocol = UDP
	case ICMPv4:
		leafICMPv4 := getICMPv4Layer(packet)
		if leafICMPv4.Info != nil {
			treeRoot.Children = append(treeRoot.Children, leafICMPv4)
		}
		treeRoot.Protocol = ICMPv4
	case ICMPv6:
		leafICMPv6 := getICMPv6Layer(packet)
		if leafICMPv6.Info != nil {
			treeRoot.Children = append(treeRoot.Children, leafICMPv6)
		}
		treeRoot.Protocol = ICMPv6
	default:
		debugPrint("Unknown protocol: " + protocol)
	}

	// VXLAN
	if treeRoot.Protocol == UDP {
		vxlanLeaf, innerEth, innerIP, innerTransport := getVXLan(packet)
		if vxlanLeaf.Info != nil {
			treeRoot.Protocol = VXLan
			treeRoot.Children = append(treeRoot.Children, vxlanLeaf)
			if innerEth.Info != nil {
				treeRoot.Children = append(treeRoot.Children, innerEth)
				if innerIP.Info != nil {
					treeRoot.Children = append(treeRoot.Children, innerIP)
					if innerTransport.Info != nil {
						treeRoot.Children = append(treeRoot.Children, innerTransport)
					}
				}
			}
		}
		return
	}

	// OpenFlow
	if treeRoot.Protocol == TCP {
		//openFlow := getOpenFlow(packet)
		openFlow := getOpenFlow(tcpContent)
		if openFlow.Info != nil {
			treeRoot.Protocol = OpenFlow
			treeRoot.Children = append(treeRoot.Children, openFlow)
		}
	}
}

func getEthernetLayer(packet gopacket.Packet) (Leaf, string, string, string) {
	etherNetLayer := packet.Layer(layers.LayerTypeEthernet)
	if etherNetLayer != nil {
		etherNet := etherNetLayer.(*layers.Ethernet)
		var leaf Leaf
		leaf.Name = "Ethernet"
		leaf.Info = "Source MAC: " + etherNet.SrcMAC.String() + "\n" + "Destination MAC: " + etherNet.DstMAC.String() + "\n" + "Ethernet type: " + etherNet.EthernetType.String()
		leaf.Hex = fmt.Sprintf("%x", etherNet.LayerContents())
		return leaf, etherNet.EthernetType.String(), etherNet.SrcMAC.String(), etherNet.DstMAC.String()
	}
	return Leaf{}, "", "", ""
}

func getIPv4Layer(packet gopacket.Packet) (Leaf, string, string, string) {
	ipv4Layer := packet.Layer(layers.LayerTypeIPv4)
	if ipv4Layer != nil {
		ipv4 := ipv4Layer.(*layers.IPv4)
		var leaf Leaf
		leaf.Name = "IPv4"
		leaf.Info = fmt.Sprintf("Version:%d\nIHL:%d\nTOS:%d\nLength:%d\nId:%d\nFlags:%d\nFragOffset:%d\nTTL:%d\nProtocol:%d\nChecksum:%d\nSource:%s\nDestination:%s", ipv4.Version, ipv4.IHL, ipv4.TOS, ipv4.Length, ipv4.Id, ipv4.Flags, ipv4.FragOffset, ipv4.TTL, ipv4.Protocol, ipv4.Checksum, ipv4.SrcIP.String(), ipv4.DstIP.String())
		leaf.Hex = fmt.Sprintf("%x", ipv4.LayerContents())
		return leaf, ipv4.Protocol.String(), ipv4.SrcIP.String(), ipv4.DstIP.String()
	}
	return Leaf{}, "", "", ""
}

func getIPv6Layer(packet gopacket.Packet) (Leaf, string, string, string) {
	ipv6Layer := packet.Layer(layers.LayerTypeIPv6)
	if ipv6Layer != nil {
		ipv6 := ipv6Layer.(*layers.IPv6)
		var leaf Leaf
		leaf.Name = "IPv6"
		leaf.Info = fmt.Sprintf("Version:%d\nLength:%d\nFlow Label:%d\nNext Header:%d\nHop Limit:%d\nSource Address:%s\nDestination Address:%s", ipv6.Version, ipv6.Length, ipv6.FlowLabel, ipv6.NextHeader, ipv6.HopLimit, ipv6.SrcIP.String(), ipv6.DstIP.String())
		leaf.Hex = fmt.Sprintf("%x", ipv6.LayerContents())
		return leaf, ipv6.NextHeader.String(), ipv6.SrcIP.String(), ipv6.DstIP.String()
	}
	return Leaf{}, "", "", ""
}

func getARP(packet gopacket.Packet) Leaf {
	arpLayer := packet.Layer(layers.LayerTypeARP)
	if arpLayer != nil {
		arp := arpLayer.(*layers.ARP)
		var leaf Leaf
		leaf.Name = "ARP"
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
		leaf.Name = "LLC"
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
		leaf.Name = "ICMPv4"
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
		leaf.Name = "ICMPv6"
		leaf.Info = fmt.Sprintf("Type:%d\nCode:%d\nChecksum:%d\nData:%s", icmpv6.TypeCode.Type(), icmpv6.TypeCode.Code(), icmpv6.Checksum, icmpv6.Payload)
		leaf.Hex = fmt.Sprintf("%x", icmpv6.LayerContents())
		return leaf
	}
	return Leaf{}
}

func getTCP(packet gopacket.Packet) (Leaf, []byte) {
	tcpLayer := packet.Layer(layers.LayerTypeTCP)
	if tcpLayer != nil {
		tcp := tcpLayer.(*layers.TCP)
		var leaf Leaf
		leaf.Name = "TCP"
		leaf.Info = fmt.Sprintf("From: %s\nTo: %s\nSeq: %d\nAck: %d\nDataOffset: %d\nFIN: %t\nSYN: %t\nRST: %t\nPSH: %t\nACK: %t\nURG: %t\nECE: %t\nCWR: %t\nNS: %t\nWindow: %d\nChecksum: %d\nUrgent: %d", tcp.SrcPort.String(), tcp.DstPort.String(), tcp.Seq, tcp.Ack, tcp.DataOffset, tcp.FIN, tcp.SYN, tcp.RST, tcp.PSH, tcp.ACK, tcp.URG, tcp.ECE, tcp.CWR, tcp.NS, tcp.Window, tcp.Checksum, tcp.Urgent)
		leaf.Hex = fmt.Sprintf("%x", tcp.LayerContents())
		return leaf, tcp.LayerPayload()
	}
	return Leaf{}, nil
}

func getUDP(packet gopacket.Packet) Leaf {
	udpLayer := packet.Layer(layers.LayerTypeUDP)
	if udpLayer != nil {
		udp := udpLayer.(*layers.UDP)
		var leaf Leaf
		leaf.Name = "UDP"
		leaf.Info = fmt.Sprintf("From: %s\nTo: %s\nLength: %d\nChecksum: %d", udp.SrcPort.String(), udp.DstPort.String(), udp.Length, udp.Checksum)
		leaf.Hex = fmt.Sprintf("%x", udp.LayerContents())
		return leaf
	}
	return Leaf{}
}

func getVXLan(packet gopacket.Packet) (VXLanLeaf, VXLanEthernet, VXLanIP, VXLanTransport Leaf) {
	vxlanLayer := packet.Layer(layers.LayerTypeVXLAN)
	if vxlanLayer != nil {
		vxlan := vxlanLayer.(*layers.VXLAN)

		var vxlanLeaf Leaf
		vxlanLeaf.Name = "VXLAN"
		vxlanLeaf.Info = fmt.Sprintf("VNI: %d\nValidIDFlag: %t\n",
			vxlan.VNI, vxlan.ValidIDFlag)
		vxlanLeaf.Hex = fmt.Sprintf("%x", vxlan.LayerContents())

		innerPacket := gopacket.NewPacket(vxlan.LayerPayload(), layers.LayerTypeEthernet, gopacket.Default)

		innerEth, _, _, _ := getEthernetLayer(innerPacket)
		innerEth.Name = "VXLAN-Inner-Ethernet"

		var innerIP Leaf
		if ipv4Layer := innerPacket.Layer(layers.LayerTypeIPv4); ipv4Layer != nil {
			innerIP, _, _, _ = getIPv4Layer(innerPacket)
			innerIP.Name = "VXLAN-Inner-IPv4"
		} else if ipv6Layer := innerPacket.Layer(layers.LayerTypeIPv6); ipv6Layer != nil {
			innerIP, _, _, _ = getIPv6Layer(innerPacket)
			innerIP.Name = "VXLAN-Inner-IPv6"
		}

		var innerTransport Leaf
		if tcpLayer := innerPacket.Layer(layers.LayerTypeTCP); tcpLayer != nil {
			innerTransport, _ = getTCP(innerPacket)
			innerTransport.Name = "VXLAN-Inner-TCP"
		} else if udpLayer := innerPacket.Layer(layers.LayerTypeUDP); udpLayer != nil {
			innerTransport = getUDP(innerPacket)
			innerTransport.Name = "VXLAN-Inner-UDP"
		} else if icmpv4Layer := innerPacket.Layer(layers.LayerTypeICMPv4); icmpv4Layer != nil {
			innerTransport = getICMPv4Layer(innerPacket)
			innerTransport.Name = "VXLAN-Inner-ICMPv4"
		} else if icmpv6Layer := innerPacket.Layer(layers.LayerTypeICMPv6); icmpv6Layer != nil {
			innerTransport = getICMPv6Layer(innerPacket)
			innerTransport.Name = "VXLAN-Inner-ICMPv6"
		}

		return vxlanLeaf, innerEth, innerIP, innerTransport
	}

	return Leaf{}, Leaf{}, Leaf{}, Leaf{}
}

func getOpenFlow(data []byte) Leaf {
	var (
		Version uint8
		Type    uint8
		Length  uint16
		Xid     uint32
	)

	tempLength := len(data)
	if tempLength < 8 {
		return Leaf{}
	}

	const (
		OFPT_0 = 0x01
		OFPT_1 = 0x02
		OFPT_2 = 0x03
		OFPT_3 = 0x04
		OFPT_4 = 0x05
		OFPT_5 = 0x06
	)

	const (
		OFPT_HELLO                    = 0x00
		OFPT_ERROR                    = 0x01
		OFPT_ECHO_REQUEST             = 0x02
		OFPT_ECHO_REPLY               = 0x03
		OFPT_EXPERIMENTER             = 0x04
		OFPT_FEATURES_REQUEST         = 0x05
		OFPT_FEATURES_REPLY           = 0x06
		OFPT_GET_CONFIG_REQUEST       = 0x07
		OFPT_GET_CONFIG_REPLY         = 0x08
		OFPT_SET_CONFIG               = 0x09
		OFPT_PACKET_IN                = 0x0a
		OFPT_FLOW_REMOVED             = 0x0b
		OFPT_PORT_STATUS              = 0x0c
		OFPT_PACKET_OUT               = 0x0d
		OFPT_FLOW_MOD                 = 0x0e
		OFPT_GROUP_MOD                = 0x0f
		OFPT_PORT_MOD                 = 0x10
		OFPT_TABLE_MOD                = 0x11
		OFPT_MULTIPART_REQUEST        = 0x12
		OFPT_MULTIPART_REPLY          = 0x13
		OFPT_BARRIER_REQUEST          = 0x14
		OFPT_BARRIER_REPLY            = 0x15
		OFPT_QUEUE_GET_CONFIG_REQUEST = 0x16
		OFPT_QUEUE_GET_CONFIG_REPLY   = 0x17
		OFPT_ROLE_REQUEST             = 0x18
		OFPT_ROLE_REPLY               = 0x19
		OFPT_GET_ASYNC_REQUEST        = 0x1a
		OFPT_GET_ASYNC_REPLY          = 0x1b
		OFPT_SET_ASYNC                = 0x1c
		OFPT_METER_MOD                = 0x1d
	)

	//读取前8个字节，判断OpenFlow版本
	Version = data[0]
	switch Version {
	case OFPT_0, OFPT_1, OFPT_2, OFPT_3:
		break
	default:
		return Leaf{}
	}

	//读取第8-15个字节，判断OpenFlow消息类型
	Type = data[1]
	switch Type {
	case OFPT_HELLO, OFPT_ERROR, OFPT_ECHO_REQUEST, OFPT_ECHO_REPLY, OFPT_EXPERIMENTER, OFPT_FEATURES_REQUEST, OFPT_FEATURES_REPLY, OFPT_GET_CONFIG_REQUEST, OFPT_GET_CONFIG_REPLY, OFPT_SET_CONFIG, OFPT_PACKET_IN, OFPT_FLOW_REMOVED, OFPT_PORT_STATUS, OFPT_PACKET_OUT, OFPT_FLOW_MOD, OFPT_GROUP_MOD, OFPT_PORT_MOD, OFPT_TABLE_MOD, OFPT_MULTIPART_REQUEST, OFPT_MULTIPART_REPLY, OFPT_BARRIER_REQUEST, OFPT_BARRIER_REPLY, OFPT_QUEUE_GET_CONFIG_REQUEST, OFPT_QUEUE_GET_CONFIG_REPLY, OFPT_ROLE_REQUEST, OFPT_ROLE_REPLY, OFPT_GET_ASYNC_REQUEST, OFPT_GET_ASYNC_REPLY, OFPT_SET_ASYNC, OFPT_METER_MOD:
		break
	default:
		//debugPrint(fmt.Sprintf("%x", Version))
		return Leaf{}
	}

	//debugPrint(2)

	//读取第16-31个字节，消息长度
	Length = binary.BigEndian.Uint16(data[2:4])
	if int(Length) != tempLength {
		debugPrint(fmt.Sprintf("Length:%d, TempLength:%d", Length, tempLength))
		return Leaf{}
	}

	//读取第32-63个字节，事务ID
	Xid = binary.BigEndian.Uint32(data[4:8])

	var leaf Leaf
	leaf.Name = "OpenFlow"
	leaf.Info = fmt.Sprintf("Version:%d\nType:%d\nLength:%d\nXid:%d", Version, Type, Length, Xid)
	leaf.Hex = fmt.Sprintf("%x", data)
	return leaf
}
