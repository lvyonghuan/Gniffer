package gniffer

func GetDataByIndex(startIndex int, endIndex int) ([]TreeRoot, int) {
	l := len(frontBuffer)
	if l == 0 {
		return nil, l - 1
	}

	if endIndex == -1 {
		if l-25 < 0 {
			readMu.RLock()
			defer readMu.RUnlock()
			return frontBuffer, l - 1
		}
		readMu.RLock()
		defer readMu.RUnlock()
		return frontBuffer[l-25:], l - 1
	}

	if startIndex < 0 {
		startIndex = 0
	}

	if startIndex >= l {
		if startIndex-25 < 0 {
			readMu.RLock()
			defer readMu.RUnlock()
			return frontBuffer, l - 1
		}
		readMu.RLock()
		defer readMu.RUnlock()
		return frontBuffer[l-25:], l - 1
	}

	if endIndex >= l {
		readMu.RLock()
		defer readMu.RUnlock()
		return frontBuffer[startIndex:], l - 1
	}

	readMu.RLock()
	defer readMu.RUnlock()
	return frontBuffer[startIndex:endIndex], l - 1
}
