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

var cpuprof cpuProfile

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

	lock(&cpuprof.lock)
	if hz > 0 {
		if cpuprof.on || cpuprof.log != nil {
			print("runtime: cannot set cpu profile rate until previous profile has finished.\n")
			unlock(&cpuprof.lock)
			return
		}

		cpuprof.on = true
		cpuprof.log = newProfBuf(1, 1<<17, 1<<14)
		hdr := [1]uint64{uint64(hz)}
		cpuprof.log.write(nil, nanotime(), hdr[:], nil)
		setcpuprofilerate(int32(hz))
	} else if cpuprof.on {
		setcpuprofilerate(0)
		cpuprof.on = false
		cpuprof.addExtra()
		cpuprof.log.close()
	}
	unlock(&cpuprof.lock)
}

// add adds the stack trace to the profile.
// It is called from signal handlers and other limited environments
// and cannot allocate memory or acquire locks that might be
// held at the time of the signal, nor can it use substantial amounts
// of stack.
//go:nowritebarrierrec
func (p *cpuProfile) add(gp *g, stk []uintptr) {
	// Simple cas-lock to coordinate with setcpuprofilerate.
	for !atomic.Cas(&prof.signalLock, 0, 1) {
		osyield()
	}

	if prof.hz != 0 { // implies cpuprof.log != nil
		if p.numExtra > 0 || p.lostExtra > 0 || p.lostAtomic > 0 {
			p.addExtra()
		}
		hdr := [1]uint64{1}
		// Note: write "knows" that the argument is &gp.labels,
		// because otherwise its write barrier behavior may not
		// be correct. See the long comment there before
		// changing the argument here.
		cpuprof.log.write(&gp.labels, nanotime(), hdr[:], stk)
	}

	atomic.Store(&prof.signalLock, 0)
}

// addNonGo adds the non-Go stack trace to the profile.
// It is called from a non-Go thread, so we cannot use much stack at all,
// nor do anything that needs a g or an m.
// In particular, we can't call cpuprof.log.write.
// Instead, we copy the stack into cpuprof.extra,
// which will be drained the next time a Go thread
// gets the signal handling event.
//go:nosplit
//go:nowritebarrierrec
func (p *cpuProfile) addNonGo(stk []uintptr) {
	// Simple cas-lock to coordinate with SetCPUProfileRate.
	// (Other calls to add or addNonGo should be blocked out
	// by the fact that only one SIGPROF can be handled by the
	// process at a time. If not, this lock will serialize those too.)
	for !atomic.Cas(&prof.signalLock, 0, 1) {
		osyield()
	}

	if cpuprof.numExtra+1+len(stk) < len(cpuprof.extra) {
		i := cpuprof.numExtra
		cpuprof.extra[i] = uintptr(1 + len(stk))
		copy(cpuprof.extra[i+1:], stk)
		cpuprof.numExtra += 1 + len(stk)
	} else {
		cpuprof.lostExtra++
	}

	atomic.Store(&prof.signalLock, 0)
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
func runtime_pprof_readProfile() ([]uint64, []unsafe.Pointer, bool) {
	lock(&cpuprof.lock)
	log := cpuprof.log
	unlock(&cpuprof.lock)
	data, tags, eof := log.read(profBufBlocking)
	if len(data) == 0 && eof {
		lock(&cpuprof.lock)
		cpuprof.log = nil
		unlock(&cpuprof.lock)
	}
	return data, tags, eof
}
