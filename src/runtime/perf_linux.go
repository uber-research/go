package runtime

import (
	"unsafe"
)

func ioctl(fd, req int32, arg uintptr) int32

//go:noescape
func perfEventOpen(attr *perfEventAttr, pid uintptr, cpu, groupFd int32, flags uintptr) int32

const perfDataPages = 2 // use 2^n data pages

func perfAttrInit(eventId cpuEvent, profConfig *cpuProfileConfig, perfAttr *perfEventAttr) {
	//perfAttr._type = _PERF_TYPE_HARDWARE //perfEventOpt[eventId]._type
	perfAttr._type = perfEventOpt[eventId]._type
	perfAttr.size = uint32(unsafe.Sizeof(*perfAttr))

	if eventId == _CPUPROF_HW_RAW {
		perfAttr.config = profConfig.rawEvent
	} else {
		//perfAttr.config = _PERF_COUNT_HW_CPU_CYCLES //perfEventOpt[eventId].config
		perfAttr.config = perfEventOpt[eventId].config
	}

	perfAttr.sample = profConfig.period
	if perfAttr.sample == 0 {
		perfAttr.read_format = _PERF_FORMAT_TOTAL_TIME_ENABLED | _PERF_FORMAT_TOTAL_TIME_RUNNING
	}
	if profConfig.isSampleIPIncluded {
		perfAttr.sample_type = _PERF_SAMPLE_IP
	}
	if profConfig.isSampleAddrIncluded {
		perfAttr.sample_type |= _PERF_SAMPLE_ADDR
	}
	if profConfig.isSampleCallchainIncluded {
		perfAttr.sample_type |= _PERF_SAMPLE_CALLCHAIN
	}
	if profConfig.isSampleThreadIDIncluded {
		perfAttr.sample_type |= _PERF_SAMPLE_TID
	}

	perfAttr.bits = 1 // the counter is disabled and will be enabled later
	// profConfig.preciseIP matches 1-1 with the Linux implementation.
	perfAttr.bits |= uint64(profConfig.preciseIP) << 15 // precise ip
	if !profConfig.isKernelIncluded {                   // don't count kernel
		perfAttr.bits |= 1 << 5
	}
	if !profConfig.isHvIncluded { // don't count hypervisor
		perfAttr.bits |= 1 << 6
	}
	//Disabled because setting this flag does not work on arm64
	//if !profConfig.isIdleIncluded { // don't count when idle
	//	perfAttr.bits |= 1 << 7
	//}
	if !profConfig.isCallchainKernelIncluded {
		perfAttr.bits |= 1 << 21
	}
	if !profConfig.isCallchainUserIncluded {
		perfAttr.bits |= 1 << 22
	}
	perfAttr.wakeup = 1 // counter overflow notifications happen after wakeup_events samples
}

func perfMmapSize() uintptr {
	perfPageSize := uint64(physPageSize)
	return uintptr(perfPageSize * (perfDataPages + 1 /* metadata page */))
}

func perfSetMmap(fd int32) unsafe.Pointer {
	size := perfMmapSize()
	mmapBuf, err := mmap(nil, size, _PROT_WRITE|_PROT_READ, _MAP_SHARED, fd, 0 /* page offset */)
	if err != 0 {
		return nil
	}
	return mmapBuf
}

func perfUnsetMmap(mmapBuf unsafe.Pointer) {
	size := perfMmapSize()
	if mmapBuf != nil {
		munmap(mmapBuf, size)
	}
}

func perfSkipNBytes(head uint64, mmapBuf *perfEventMmapPage, n uint64) {
	tail := mmapBuf.data_tail
	remains := head - tail
	if n > remains {
		n = remains
	}
	mmapBuf.data_tail += n
}

func perfSkipRecord(head uint64, mmapBuf *perfEventMmapPage, hdr *perfEventHeader) {
	if mmapBuf == nil {
		return
	}
	remains := uint64(hdr.size) - uint64(unsafe.Sizeof(*hdr))
	if remains > 0 {
		perfSkipNBytes(head, mmapBuf, remains)
	}
}

func perfSkipAll(head uint64, mmapBuf *perfEventMmapPage) {
	if mmapBuf == nil {
		return
	}
	tail := mmapBuf.data_tail
	remains := head - tail
	if remains > 0 {
		mmapBuf.data_tail += remains
	}
}

func perfReadNbytes(head uint64, mmapBuf *perfEventMmapPage, buf unsafe.Pointer, n uint64) bool {
	if mmapBuf == nil {
		return false
	}
	perfPageSize := uint64(physPageSize)
	perfPageMask := perfPageSize*perfDataPages - 1
	// front of the circular data buffer
	data := unsafe.Pointer(uintptr(unsafe.Pointer(mmapBuf)) + uintptr(perfPageSize))
	tail := mmapBuf.data_tail
	// compute bytes available in the circular buffer
	byteAvailable := head - tail
	if n > byteAvailable {
		return false
	}
	// compute offset of tail in the circular buffer
	tail &= perfPageMask
	bytesAtRight := (perfPageMask + 1) - tail
	// bytes to copy to the right of tail
	var right uint64
	if bytesAtRight < n {
		right = bytesAtRight
	} else {
		right = n
	}
	// copy bytes from tail position
	memmove(buf, unsafe.Pointer(uintptr(data)+uintptr(tail)), uintptr(right))
	// if necessary, wrap and continue copy from left edge of buf
	if n > right {
		left := n - right
		memmove(unsafe.Pointer(uintptr(buf)+uintptr(right)), data, uintptr(left))
	}
	// update tail after consuming n bytes
	mmapBuf.data_tail += n
	return true
}

func perfReadHeader(head uint64, mmapBuf *perfEventMmapPage, hdr *perfEventHeader) bool {
	return perfReadNbytes(head, mmapBuf, unsafe.Pointer(hdr), uint64(unsafe.Sizeof(*hdr)))
}

// The order where values are read has to match the mmap ring buffer layout
func perfRecordSample(head uint64, mmapBuf *perfEventMmapPage, profConfig *cpuProfileConfig, sampleData *perfSampleData) {
	if profConfig.isSampleIPIncluded {
		perfReadNbytes(head, mmapBuf, unsafe.Pointer(&(sampleData.ip)), uint64(unsafe.Sizeof(sampleData.ip)))
	}
	if profConfig.isSampleThreadIDIncluded {
		perfReadNbytes(head, mmapBuf, unsafe.Pointer(&(sampleData.pid)), uint64(unsafe.Sizeof(sampleData.pid)))
		perfReadNbytes(head, mmapBuf, unsafe.Pointer(&(sampleData.tid)), uint64(unsafe.Sizeof(sampleData.tid)))
	}
	if profConfig.isSampleAddrIncluded {
		perfReadNbytes(head, mmapBuf, unsafe.Pointer(&(sampleData.addr)), uint64(unsafe.Sizeof(sampleData.addr)))
	}
}

func perfStartCounter(fd int32) bool {
	err := ioctl(fd, _PERF_EVENT_IOC_ENABLE, uintptr(0))
	if err != 0 {
		println("Failed to enable the event count")
		return false
	}
	return true
}

func perfStopCounter(fd int32) bool {
	err := ioctl(fd, _PERF_EVENT_IOC_DISABLE, uintptr(0))
	if err != 0 {
		println("Failed to disable the event count")
		return false
	}
	return true
}

func perfResetCounter(fd int32) bool {
	err := ioctl(fd, _PERF_EVENT_IOC_RESET, uintptr(0))
	if err != 0 {
		println("Failed to reset the event count")
		return false
	}
	return true
}

func perfReadCounter(fd int32, val *uint64) bool {
	return read(fd, unsafe.Pointer(val), int32(unsafe.Sizeof(*val))) != -1
}

func perfConsumeSampleData(mmapBuf *perfEventMmapPage, profConfig *cpuProfileConfig) {
	if mmapBuf == nil || profConfig == nil {
		return
	}

	head := mmapBuf.data_head
	rmb()
	for {
		tail := mmapBuf.data_tail
		remains := head - tail
		if remains <= 0 {
			break
		}
		var hdr perfEventHeader
		if remains < uint64(unsafe.Sizeof(hdr)) {
			perfSkipAll(head, mmapBuf)
			break
		}
		if !perfReadHeader(head, mmapBuf, &hdr) {
			break
		}
		if hdr.size == 0 {
			perfSkipAll(head, mmapBuf)
			break
		}
		switch hdr._type {
		case _PERF_RECORD_SAMPLE:
			var sampleData perfSampleData
			sampleData.isPreciseIP = (hdr.misc & _PERF_RECORD_MISC_EXACT_IP) != 0
			perfRecordSample(head, mmapBuf, profConfig, &sampleData)
		default:
			perfSkipRecord(head, mmapBuf, &hdr)
		}
	}
}
