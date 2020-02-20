package runtime

// Convert platform-agnostic pmu events to Linux perf events
var perfEventOpt = map[cpuEvent]struct {
	_type  uint32 // type of event
	config uint64 // event
}{
	_CPUPROF_HW_CPU_CYCLES:          {_PERF_TYPE_HARDWARE, _PERF_COUNT_HW_CPU_CYCLES},
	_CPUPROF_HW_INSTRUCTIONS:        {_PERF_TYPE_HARDWARE, _PERF_COUNT_HW_INSTRUCTIONS},
	_CPUPROF_HW_CACHE_REFERENCES:    {_PERF_TYPE_HARDWARE, _PERF_COUNT_HW_CACHE_REFERENCES},
	_CPUPROF_HW_CACHE_MISSES:        {_PERF_TYPE_HARDWARE, _PERF_COUNT_HW_CACHE_MISSES},
	_CPUPROF_HW_BRANCH_INSTRUCTIONS: {_PERF_TYPE_HARDWARE, _PERF_COUNT_HW_BRANCH_INSTRUCTIONS},
	_CPUPROF_HW_BRANCH_MISSES:       {_PERF_TYPE_HARDWARE, _PERF_COUNT_HW_BRANCH_MISSES},
	_CPUPROF_HW_RAW:                 {_PERF_TYPE_RAW, 0 /* will not be used */},
	// TODO: add more perf events
}

type perfEventAttr struct {
	_type              uint32
	size               uint32
	config             uint64
	sample             uint64
	sample_type        uint64
	read_format        uint64
	bits               uint64
	wakeup             uint32
	bp_type            uint32
	ext1               uint64
	ext2               uint64
	branch_sample_type uint64
	sample_regs_user   uint64
	sample_stack_user  uint32
	clockid            int32
	sample_regs_intr   uint64
	aux_watermark      uint32
	sample_max_stack   uint16
	_                  uint16
}

type perfEventMmapPage struct {
	version        uint32
	compat_version uint32
	lock           uint32
	index          uint32
	offset         int64
	time_enabled   uint64
	time_running   uint64
	capabilities   uint64
	pmc_width      uint16
	time_shift     uint16
	time_mult      uint32
	time_offset    uint64
	time_zero      uint64
	size           uint32
	_              [948]uint8
	data_head      uint64
	data_tail      uint64
	data_offset    uint64
	data_size      uint64
	aux_head       uint64
	aux_tail       uint64
	aux_offset     uint64
	aux_size       uint64
}

type perfEventHeader struct {
	_type uint32
	misc  uint16
	size  uint16
}

// The order where values are saved in a sample has to match the mmap ring buffer layout
type perfSampleData struct {
	ip   uint64 // if _PERF_SAMPLE_IP
	pid  uint32 // if _PERF_SAMPLE_TID
	tid  uint32 // if _PERF_SAMPLE_TID
	addr uint64 // if _PERF_SAMPLE_ADDR
	// TODO: More fields can be added in order if needed

	/*********** auxiliary fields ***********/
	isPreciseIP bool // whether the obtained ip is precise or not
}
