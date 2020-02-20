// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// CPU profiling.
//
// The signal handler for the profiling clock tick adds a new stack trace
// to a log of recent traces. The log is read by a user goroutine that
// turns it into formatted profile data. If the reader does not keep up
// with the log, those writes will be recorded as a count of lost records.
// The actual profile buffer is in profbuf.go.

package runtime

import (
	"runtime/internal/atomic"
	"runtime/internal/sys"
	"unsafe"
)

const maxCPUProfStack = 64

//go:linkname cpuEvent runtime/pprof.cpuEvent
type cpuEvent int32

const (
	_CPUPROF_OS_TIMER, _CPUPROF_FIRST_EVENT cpuEvent = iota, iota
	_CPUPROF_HW_CPU_CYCLES, _CPUPROF_FIRST_PMU_EVENT
	_CPUPROF_HW_INSTRUCTIONS cpuEvent = iota
	_CPUPROF_HW_CACHE_REFERENCES
	_CPUPROF_HW_CACHE_MISSES
	_CPUPROF_HW_BRANCH_INSTRUCTIONS
	_CPUPROF_HW_BRANCH_MISSES
	_CPUPROF_HW_RAW
	_CPUPROF_EVENTS_MAX
	_CPUPROF_LAST_EVENT cpuEvent = _CPUPROF_EVENTS_MAX - 1
)

//go:linkname profilePCPrecision runtime/pprof.profilePCPrecision
type profilePCPrecision uint8

const (
	_CPUPROF_IP_ARBITRARY_SKID profilePCPrecision = iota
	_CPUPROF_IP_CONSTANT_SKID
	_CPUPROF_IP_SUGGEST_NO_SKID
	_CPUPROF_IP_NO_SKID
)

// cpuProfileConfig holds different settings under which CPU samples can be produced.
// Not all fields are in use yet.
// hz defines the rate of sampling (used only for OS timer-based sampling).
// period is complementory to hz. The period defines the interval (the number of events that elase) between generating two sampling interrupts.
// rawEvent is an opaque number that is passed down to CPU to specify what event to sample.
// The raw event is CPU vendor and version specific.
// preciseIP is one of the following value:
//    CPUPROF_IP_ARBITRARY_SKID: no skid constaint from when the sample occurs to when the interrupt is generated,
//    CPUPROF_IP_CONSTANT_SKID: a constant skid between a sample and the corresponding interrupt,
//    CPUPROF_IP_SUGGEST_NO_SKID: request zero skid between a sample and the corresponding interrupt, but no guarantee,
//    CPUPROF_IP_NO_SKID: demand no skid between a sample and the corresponding interrupt.
// isSampleIPIncluded: include the instuction pointer that caused the sample to occur.
// isSampleThreadIDIncluded: include the thread id in the sample.
// isSampleAddrIncluded: include the memory address accessed at the time of generating the sample.
// isKernelIncluded: count the events in the kernel mode.
// isHvIncluded: count the events in the hypervisor mode.
// isHvIncluded: include the kernel call chain at the time of the sample.
// isIdleIncluded: count when the CPU is running the idle task.
// isSampleCallchainIncluded: include the entire call chain seen at the time of the sample.
// isCallchainKernelIncluded: include the kernel call chain seen at the time of the sample.
// isCallchainUserIncluded: include the user call chain seen at the time of the sample.
//
//go:linkname cpuProfileConfig runtime/pprof.cpuProfileConfig
type cpuProfileConfig struct {
	hz                        uint64
	period                    uint64
	rawEvent                  uint64
	preciseIP                 profilePCPrecision
	isSampleIPIncluded        bool
	isSampleThreadIDIncluded  bool
	isSampleAddrIncluded      bool
	isKernelIncluded          bool
	isHvIncluded              bool
	isIdleIncluded            bool
	isSampleCallchainIncluded bool
	isCallchainKernelIncluded bool
	isCallchainUserIncluded   bool
}

type cpuProfile struct {
	lock mutex
	on   bool     // profiling is on
	log  *profBuf // profile events written here

	// extra holds extra stacks accumulated in addNonGo
	// corresponding to profiling signals arriving on
	// non-Go-created threads. Those stacks are written
	// to log the next time a normal Go thread gets the
	// signal handler.
	// Assuming the stacks are 2 words each (we don't get
	// a full traceback from those threads), plus one word
	// size for framing, 100 Hz profiling would generate
	// 300 words per second.
	// Hopefully a normal Go thread will get the profiling
	// signal at least once every few seconds.
	extra      [1000]uintptr
	numExtra   int
	lostExtra  uint64 // count of frames lost because extra is full
	lostAtomic uint64 // count of frames lost because of being in atomic64 on mips/arm; updated racily
}

var cpuprof [_CPUPROF_EVENTS_MAX]cpuProfile

// SetCPUProfileRate sets the CPU profiling rate to hz samples per second.
// If hz <= 0, SetCPUProfileRate turns off profiling.
// If the profiler is on, the rate cannot be changed without first turning it off.
//
// Most clients should use the runtime/pprof package or
// the testing package's -test.cpuprofile flag instead of calling
// SetCPUProfileRate directly.
func SetCPUProfileRate(hz int) {
	// Clamp hz to something reasonable.
	if hz < 0 {
		hz = 0
	}
	if hz > 1000000 {
		hz = 1000000
	}

	if hz > 0 {
		var profConfig cpuProfileConfig
		profConfig.hz = uint64(hz)
		runtime_pprof_setCPUProfileConfig(_CPUPROF_OS_TIMER, &profConfig)
	} else {
		runtime_pprof_setCPUProfileConfig(_CPUPROF_OS_TIMER, nil)
	}
}

func sanitizeCPUProfileConfig(profConfig *cpuProfileConfig) {
	if profConfig == nil {
		return
	}
	profConfig.preciseIP = _CPUPROF_IP_ARBITRARY_SKID
	profConfig.isSampleIPIncluded = false
	profConfig.isSampleThreadIDIncluded = false
	profConfig.isSampleAddrIncluded = false
	profConfig.isSampleCallchainIncluded = false
	profConfig.isKernelIncluded = false
	profConfig.isHvIncluded = false
	profConfig.isIdleIncluded = false
	profConfig.isCallchainKernelIncluded = false
	profConfig.isCallchainUserIncluded = false
}

// setCPUProfileConfig, provided to runtime/pprof, enables/disables CPU profiling for a specified CPU event.
// Profiling cannot be enabled if it is already enabled.
// eventId: specifies the event to enable/disable. eventId can be one of the following values:
//      _CPUPROF_OS_TIMER, _CPUPROF_HW_CPU_CYCLES, _CPUPROF_HW_INSTRUCTIONS, _CPUPROF_HW_CACHE_REFERENCES,
//      _CPUPROF_HW_CACHE_MISSES, CPUPROF_HW_CACHE_LL_READ_ACCESSES, CPUPROF_HW_CACHE_LL_READ_MISSES, _CPUPROF_HW_RAW
// profConfig: provides additional configurations when enabling the specified event.
//             A nil profConfig results in disabling the said event.
// TODO: should we make this function return an error?
//
//go:linkname runtime_pprof_setCPUProfileConfig runtime/pprof.setCPUProfileConfig
func runtime_pprof_setCPUProfileConfig(eventId cpuEvent, profConfig *cpuProfileConfig) {
	if eventId >= _CPUPROF_EVENTS_MAX {
		return
	}

	lock(&cpuprof[eventId].lock)
	defer unlock(&cpuprof[eventId].lock)
	if profConfig != nil {
		if cpuprof[eventId].on || cpuprof[eventId].log != nil {
			print("runtime: cannot set cpu profile config until previous profile has finished.\n")
			return
		}

		cpuprof[eventId].on = true
		// Enlarging the buffer words and tags reduces the number of samples lost at the cost of larger amounts of memory
		cpuprof[eventId].log = newProfBuf( /* header size */ 1 /* buffer words */, 1<<17 /* tags */, 1<<14)
		// OS timer profiling provides the sampling rate (sample/sec), whereas the other PMU-based events provide
		// sampling interval (aka period), which is the the number of events to elapse before a sample is triggered.
		// The latter is called as "event-based sampling". In event-based sampling, the overhead is proportional to the
		// number of events; no events imples no overhead.
		// On Linux-based systems perf_event_open() allows configuring PMU-events in a "Hz" mode; but that is for later.
		if eventId == _CPUPROF_OS_TIMER {
			hdr := [1]uint64{profConfig.hz}
			cpuprof[eventId].log.write(nil, nanotime(), hdr[:], nil)
		} else {
			hdr := [1]uint64{profConfig.period}
			cpuprof[eventId].log.write(nil, nanotime(), hdr[:], nil)
		}
		// Take a copy of the profConfig passed by the user, so that the runtime functions are not affected
		// if the user code changes the attributes.
		cfg := make([]cpuProfileConfig, 1, 1)
		cfg[0] = *profConfig
		sanitizeCPUProfileConfig(&cfg[0])
		setcpuprofileconfig(eventId, &cfg[0])
	} else if cpuprof[eventId].on {
		setcpuprofileconfig(eventId, nil)
		cpuprof[eventId].on = false
		cpuprof[eventId].addExtra()
		cpuprof[eventId].log.close()
	}
}

// add adds the stack trace to the profile.
// It is called from signal handlers and other limited environments
// and cannot allocate memory or acquire locks that might be
// held at the time of the signal, nor can it use substantial amounts
// of stack.
//go:nowritebarrierrec
func (p *cpuProfile) add(gp *g, stk []uintptr, eventId cpuEvent) {
	profCfg := &prof[eventId]
	for !atomic.Cas(&signalLock, 0, 1) {
		osyield()
	}
	if profCfg.config != nil { // implies cpuprof[eventId].log != nil
		if p.numExtra > 0 || p.lostExtra > 0 || p.lostAtomic > 0 {
			p.addExtra()
		}
		hdr := [1]uint64{1}
		// Note: write "knows" that the argument is &gp.labels,
		// because otherwise its write barrier behavior may not
		// be correct. See the long comment there before
		// changing the argument here.
		cpuprof[eventId].log.write(&gp.labels, nanotime(), hdr[:], stk)
	}
	atomic.Store(&signalLock, 0)
}

// addNonGo adds the non-Go stack trace to the profile.
// It is called from a non-Go thread, so we cannot use much stack at all,
// nor do anything that needs a g or an m.
// In particular, we can't call cpuprof[id].log.write.
// Instead, we copy the stack into cpuprof[id].extra,
// which will be drained the next time a Go thread
// gets the signal handling event.
//go:nosplit
//go:nowritebarrierrec
func (p *cpuProfile) addNonGo(stk []uintptr, eventId cpuEvent) {
	// Simple cas-lock to coordinate with SetCPUProfileRate.
	// (Other calls to add or addNonGo should be blocked out
	// by the fact that only one SIGPROF can be handled by the
	// process at a time. If not, this lock will serialize those too.)
	for !atomic.Cas(&signalLock, 0, 1) {
		osyield()
	}
	prof := &cpuprof[eventId]
	if prof.numExtra+1+len(stk) < len(prof.extra) {
		i := prof.numExtra
		prof.extra[i] = uintptr(1 + len(stk))
		copy(prof.extra[i+1:], stk)
		prof.numExtra += 1 + len(stk)
	} else {
		prof.lostExtra++
	}
	atomic.Store(&signalLock, 0)
}

// addExtra adds the "extra" profiling events,
// queued by addNonGo, to the profile log.
// addExtra is called either from a signal handler on a Go thread
// or from an ordinary goroutine; either way it can use stack
// and has a g. The world may be stopped, though.
func (p *cpuProfile) addExtra() {
	// Copy accumulated non-Go profile events.
	hdr := [1]uint64{1}
	for i := 0; i < p.numExtra; {
		p.log.write(nil, 0, hdr[:], p.extra[i+1:i+int(p.extra[i])])
		i += int(p.extra[i])
	}
	p.numExtra = 0

	// Report any lost events.
	if p.lostExtra > 0 {
		hdr := [1]uint64{p.lostExtra}
		lostStk := [2]uintptr{
			funcPC(_LostExternalCode) + sys.PCQuantum,
			funcPC(_ExternalCode) + sys.PCQuantum,
		}
		p.log.write(nil, 0, hdr[:], lostStk[:])
		p.lostExtra = 0
	}

	if p.lostAtomic > 0 {
		hdr := [1]uint64{p.lostAtomic}
		lostStk := [2]uintptr{
			funcPC(_LostSIGPROFDuringAtomic64) + sys.PCQuantum,
			funcPC(_System) + sys.PCQuantum,
		}
		p.log.write(nil, 0, hdr[:], lostStk[:])
		p.lostAtomic = 0
	}

}

// CPUProfile panics.
// It formerly provided raw access to chunks of
// a pprof-format profile generated by the runtime.
// The details of generating that format have changed,
// so this functionality has been removed.
//
// Deprecated: Use the runtime/pprof package,
// or the handlers in the net/http/pprof package,
// or the testing package's -test.cpuprofile flag instead.
func CPUProfile() []byte {
	panic("CPUProfile no longer available")
}

//go:linkname runtime_pprof_runtime_cyclesPerSecond runtime/pprof.runtime_cyclesPerSecond
func runtime_pprof_runtime_cyclesPerSecond() int64 {
	return tickspersecond()
}

// readProfile, provided to runtime/pprof, returns the next chunk of
// binary CPU profiling stack trace data, blocking until data is available.
// If profiling is turned off and all the profile data accumulated while it was
// on has been returned, readProfile returns eof=true.
// The caller must save the returned data and tags before calling readProfile again.
//
//go:linkname runtime_pprof_readProfile runtime/pprof.readProfile
func runtime_pprof_readProfile(eventId cpuEvent) ([]uint64, []unsafe.Pointer, bool) {
	prof := &cpuprof[eventId]
	lock(&prof.lock)
	log := prof.log
	unlock(&prof.lock)
	data, tags, eof := log.read(profBufBlocking)
	if len(data) == 0 && eof {
		lock(&prof.lock)
		prof.log = nil
		unlock(&prof.lock)
	}
	return data, tags, eof
}
