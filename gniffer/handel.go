package gniffer

import (
	"sort"
	"sync"
	"time"
)

//去他妈的并发BUG//

var beforeHandelBuffer *[]TreeRoot
var frontBuffer []TreeRoot
var readMu sync.RWMutex

var SortType = ""
var FilterType = NoFilter

//处理数据，呈现给传递前端的格式

const (
	ARP      = "ARP"
	IPv4     = "IPv4"
	IPv6     = "IPv6"
	TCP      = "TCP"
	UDP      = "UDP"
	LLC      = "LLC"
	ICMPv4   = "ICMPv4"
	ICMPv6   = "ICMPv6"
	NoFilter = "NoFilter"

	ID          = "id"
	Time        = "time"
	Source      = "source"
	Destination = "destination"
	Protocol    = "protocol"
	Length      = "length"
)

func HandelData() {
	ticker := time.NewTicker(750 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			filterBuffer := filterData(FilterType)
			//sortBuffer := sortData(SortType, filterBuffer)
			readMu.Lock()
			frontBuffer = filterBuffer
			readMu.Unlock()
		}
	}
}

func filterData(protocolType string) []TreeRoot {
	if beforeHandelBuffer == nil {
		return nil
	}

	if protocolType == NoFilter {
		return *beforeHandelBuffer
	} else {
		var filterBuffer []TreeRoot
		for _, data := range *beforeHandelBuffer {
			if data.Protocol == protocolType {
				filterBuffer = append(filterBuffer, data)
			}
		}
		return filterBuffer
	}
}

func sortData(sortType string, filterBuffer []TreeRoot) []TreeRoot {
	switch sortType {
	case ID:
		return sortByID(filterBuffer)
	case Time:
		return sortDataByTime(filterBuffer)
	case Source:
		return sortDataBySource(filterBuffer)
	case Destination:
		return sortDataByDestination(filterBuffer)
	case Protocol:
		return sortDataByProtocol(filterBuffer)
	case Length:
		return sortDataByLength(filterBuffer)
	default:
		return sortByID(filterBuffer)
	}
}

func sortByID(data []TreeRoot) []TreeRoot {
	sort.Slice(data, func(i, j int) bool {
		return data[i].ID < data[j].ID
	})

	return data
}

func sortDataByTime(data []TreeRoot) []TreeRoot {
	sort.Slice(data, func(i, j int) bool {
		return data[i].time.Before(data[j].time)
	})

	return data
}

func sortDataBySource(data []TreeRoot) []TreeRoot {
	sort.Slice(data, func(i, j int) bool {
		return data[i].Source < data[j].Source
	})

	return data
}

func sortDataByDestination(data []TreeRoot) []TreeRoot {
	sort.Slice(data, func(i, j int) bool {
		return data[i].Destination < data[j].Destination
	})

	return data
}

func sortDataByProtocol(data []TreeRoot) []TreeRoot {
	sort.Slice(data, func(i, j int) bool {
		return data[i].Protocol < data[j].Protocol
	})

	return data
}

func sortDataByLength(data []TreeRoot) []TreeRoot {
	sort.Slice(data, func(i, j int) bool {
		return data[i].Length < data[j].Length
	})

	return data
}
