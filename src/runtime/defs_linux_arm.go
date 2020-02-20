package runtime

// Constants
const (
	_EINTR  = 0x4
	_ENOMEM = 0xc
	_EAGAIN = 0xb
	_ENOSYS = 0x26

	_PROT_NONE  = 0
	_PROT_READ  = 0x1
	_PROT_WRITE = 0x2
	_PROT_EXEC  = 0x4

	_MAP_SHARED          = 0x1
	_MAP_PRIVATE         = 0x2
	_MAP_SHARED_VALIDATE = 0x3
	_MAP_FIXED           = 0x10
	_MAP_ANON            = 0x20

	_MADV_DONTNEED   = 0x4
	_MADV_FREE       = 0x8
	_MADV_HUGEPAGE   = 0xe
	_MADV_NOHUGEPAGE = 0xf

	_SA_RESTART  = 0x10000000
	_SA_ONSTACK  = 0x8000000
	_SA_RESTORER = 0 // unused on ARM
	_SA_SIGINFO  = 0x4
	_SIGHUP      = 0x1
	_SIGINT      = 0x2
	_SIGQUIT     = 0x3
	_SIGILL      = 0x4
	_SIGTRAP     = 0x5
	_SIGABRT     = 0x6
	_SIGBUS      = 0x7
	_SIGFPE      = 0x8
	_SIGKILL     = 0x9
	_SIGUSR1     = 0xa
	_SIGSEGV     = 0xb
	_SIGUSR2     = 0xc
	_SIGPIPE     = 0xd
	_SIGALRM     = 0xe
	_SIGSTKFLT   = 0x10
	_SIGCHLD     = 0x11
	_SIGCONT     = 0x12
	_SIGSTOP     = 0x13
	_SIGTSTP     = 0x14
	_SIGTTIN     = 0x15
	_SIGTTOU     = 0x16
	_SIGURG      = 0x17
	_SIGXCPU     = 0x18
	_SIGXFSZ     = 0x19
	_SIGVTALRM   = 0x1a
	_SIGPROF     = 0x1b
	_SIGWINCH    = 0x1c
	_SIGIO       = 0x1d
	_SIGPWR      = 0x1e
	_SIGSYS      = 0x1f

	_FPE_INTDIV     = 0x1
	_FPE_INTOVF     = 0x2
	_FPE_FLTDIV     = 0x3
	_FPE_FLTOVF     = 0x4
	_FPE_FLTUND     = 0x5
	_FPE_FLTRES     = 0x6
	_FPE_FLTINV     = 0x7
	_FPE_FLTSUB     = 0x8
	_BUS_ADRALN     = 0x1
	_BUS_ADRERR     = 0x2
	_BUS_OBJERR     = 0x3
	_SEGV_MAPERR    = 0x1
	_SEGV_ACCERR    = 0x2
	_ITIMER_REAL    = 0
	_ITIMER_PROF    = 0x2
	_ITIMER_VIRTUAL = 0x1
	_O_RDONLY       = 0
	_O_NONBLOCK     = 0x800
	_O_CLOEXEC      = 0x80000

	_EPOLLIN       = 0x1
	_EPOLLOUT      = 0x4
	_EPOLLERR      = 0x8
	_EPOLLHUP      = 0x10
	_EPOLLRDHUP    = 0x2000
	_EPOLLET       = 0x80000000
	_EPOLL_CLOEXEC = 0x80000
	_EPOLL_CTL_ADD = 0x1
	_EPOLL_CTL_DEL = 0x2
	_EPOLL_CTL_MOD = 0x3

	_AF_UNIX    = 0x1
	_SOCK_DGRAM = 0x2

	_F_OWNER_TID = 0x0
	_F_GETFL     = 0x3
	_F_SETFL     = 0x4
	_F_SETSIG    = 0xa
	_F_SETOWN_EX = 0xf
	_O_ASYNC     = 0x2000
)

// The replication is because constants could be different on different architectures
const (
	_PERF_TYPE_HARDWARE   = 0x0
	_PERF_TYPE_SOFTWARE   = 0x1
	_PERF_TYPE_TRACEPOINT = 0x2
	_PERF_TYPE_HW_CACHE   = 0x3
	_PERF_TYPE_RAW        = 0x4
	_PERF_TYPE_BREAKPOINT = 0x5

	_PERF_COUNT_HW_CPU_CYCLES              = 0x0
	_PERF_COUNT_HW_INSTRUCTIONS            = 0x1
	_PERF_COUNT_HW_CACHE_REFERENCES        = 0x2
	_PERF_COUNT_HW_CACHE_MISSES            = 0x3
	_PERF_COUNT_HW_BRANCH_INSTRUCTIONS     = 0x4
	_PERF_COUNT_HW_BRANCH_MISSES           = 0x5
	_PERF_COUNT_HW_BUS_CYCLES              = 0x6
	_PERF_COUNT_HW_STALLED_CYCLES_FRONTEND = 0x7
	_PERF_COUNT_HW_STALLED_CYCLES_BACKEND  = 0x8
	_PERF_COUNT_HW_REF_CPU_CYCLES          = 0x9

	_PERF_COUNT_HW_CACHE_L1D  = 0x0
	_PERF_COUNT_HW_CACHE_L1I  = 0x1
	_PERF_COUNT_HW_CACHE_LL   = 0x2
	_PERF_COUNT_HW_CACHE_DTLB = 0x3
	_PERF_COUNT_HW_CACHE_ITLB = 0x4
	_PERF_COUNT_HW_CACHE_BPU  = 0x5
	_PERF_COUNT_HW_CACHE_NODE = 0x6

	_PERF_COUNT_HW_CACHE_OP_READ     = 0x0
	_PERF_COUNT_HW_CACHE_OP_WRITE    = 0x1
	_PERF_COUNT_HW_CACHE_OP_PREFETCH = 0x2

	_PERF_COUNT_HW_CACHE_RESULT_ACCESS = 0x0
	_PERF_COUNT_HW_CACHE_RESULT_MISS   = 0x1

	_PERF_COUNT_SW_CPU_CLOCK        = 0x0
	_PERF_COUNT_SW_TASK_CLOCK       = 0x1
	_PERF_COUNT_SW_PAGE_FAULTS      = 0x2
	_PERF_COUNT_SW_CONTEXT_SWITCHES = 0x3
	_PERF_COUNT_SW_CPU_MIGRATIONS   = 0x4
	_PERF_COUNT_SW_PAGE_FAULTS_MIN  = 0x5
	_PERF_COUNT_SW_PAGE_FAULTS_MAJ  = 0x6
	_PERF_COUNT_SW_ALIGNMENT_FAULTS = 0x7
	_PERF_COUNT_SW_EMULATION_FAULTS = 0x8
	_PERF_COUNT_SW_DUMMY            = 0x9
	_PERF_COUNT_SW_BPF_OUTPUT       = 0xa

	_PERF_SAMPLE_IP           = 0x1
	_PERF_SAMPLE_TID          = 0x2
	_PERF_SAMPLE_TIME         = 0x4
	_PERF_SAMPLE_ADDR         = 0x8
	_PERF_SAMPLE_READ         = 0x10
	_PERF_SAMPLE_CALLCHAIN    = 0x20
	_PERF_SAMPLE_ID           = 0x40
	_PERF_SAMPLE_CPU          = 0x80
	_PERF_SAMPLE_PERIOD       = 0x100
	_PERF_SAMPLE_STREAM_ID    = 0x200
	_PERF_SAMPLE_RAW          = 0x400
	_PERF_SAMPLE_BRANCH_STACK = 0x800

	_PERF_SAMPLE_BRANCH_USER       = 0x1
	_PERF_SAMPLE_BRANCH_KERNEL     = 0x2
	_PERF_SAMPLE_BRANCH_HV         = 0x4
	_PERF_SAMPLE_BRANCH_ANY        = 0x8
	_PERF_SAMPLE_BRANCH_ANY_CALL   = 0x10
	_PERF_SAMPLE_BRANCH_ANY_RETURN = 0x20
	_PERF_SAMPLE_BRANCH_IND_CALL   = 0x40
	_PERF_SAMPLE_BRANCH_ABORT_TX   = 0x80
	_PERF_SAMPLE_BRANCH_IN_TX      = 0x100
	_PERF_SAMPLE_BRANCH_NO_TX      = 0x200
	_PERF_SAMPLE_BRANCH_COND       = 0x400
	_PERF_SAMPLE_BRANCH_CALL_STACK = 0x800
	_PERF_SAMPLE_BRANCH_IND_JUMP   = 0x1000
	_PERF_SAMPLE_BRANCH_CALL       = 0x2000
	_PERF_SAMPLE_BRANCH_NO_FLAGS   = 0x4000
	_PERF_SAMPLE_BRANCH_NO_CYCLES  = 0x8000
	_PERF_SAMPLE_BRANCH_TYPE_SAVE  = 0x10000

	_PERF_FORMAT_TOTAL_TIME_ENABLED = 0x1
	_PERF_FORMAT_TOTAL_TIME_RUNNING = 0x2
	_PERF_FORMAT_ID                 = 0x4
	_PERF_FORMAT_GROUP              = 0x8

	_PERF_RECORD_MISC_EXACT_IP = 0x4000

	_PERF_RECORD_MMAP            = 0x1
	_PERF_RECORD_LOST            = 0x2
	_PERF_RECORD_COMM            = 0x3
	_PERF_RECORD_EXIT            = 0x4
	_PERF_RECORD_THROTTLE        = 0x5
	_PERF_RECORD_UNTHROTTLE      = 0x6
	_PERF_RECORD_FORK            = 0x7
	_PERF_RECORD_READ            = 0x8
	_PERF_RECORD_SAMPLE          = 0x9
	_PERF_RECORD_MMAP2           = 0xa
	_PERF_RECORD_AUX             = 0xb
	_PERF_RECORD_ITRACE_START    = 0xc
	_PERF_RECORD_LOST_SAMPLES    = 0xd
	_PERF_RECORD_SWITCH          = 0xe
	_PERF_RECORD_SWITCH_CPU_WIDE = 0xf
	_PERF_RECORD_NAMESPACES      = 0x10

	_PERF_CONTEXT_HV     = -0x20
	_PERF_CONTEXT_KERNEL = -0x80
	_PERF_CONTEXT_USER   = -0x200

	_PERF_CONTEXT_GUEST        = -0x800
	_PERF_CONTEXT_GUEST_KERNEL = -0x880
	_PERF_CONTEXT_GUEST_USER   = -0xa00

	_PERF_FLAG_FD_NO_GROUP = 0x1
	_PERF_FLAG_FD_OUTPUT   = 0x2
	_PERF_FLAG_PID_CGROUP  = 0x4
	_PERF_FLAG_FD_CLOEXEC  = 0x8

	_PERF_EVENT_IOC_DISABLE           = 0x2401
	_PERF_EVENT_IOC_ENABLE            = 0x2400
	_PERF_EVENT_IOC_ID                = 0x80042407
	_PERF_EVENT_IOC_MODIFY_ATTRIBUTES = 0x4004240b
	_PERF_EVENT_IOC_PAUSE_OUTPUT      = 0x40042409
	_PERF_EVENT_IOC_PERIOD            = 0x40082404
	_PERF_EVENT_IOC_QUERY_BPF         = 0xc004240a
	_PERF_EVENT_IOC_REFRESH           = 0x2402
	_PERF_EVENT_IOC_RESET             = 0x2403
	_PERF_EVENT_IOC_SET_BPF           = 0x40042408
	_PERF_EVENT_IOC_SET_FILTER        = 0x40042406
	_PERF_EVENT_IOC_SET_OUTPUT        = 0x2405
)

type timespec struct {
	tv_sec  int32
	tv_nsec int32
}

//go:nosplit
func (ts *timespec) setNsec(ns int64) {
	ts.tv_sec = timediv(ns, 1e9, &ts.tv_nsec)
}

type stackt struct {
	ss_sp    *byte
	ss_flags int32
	ss_size  uintptr
}

type sigcontext struct {
	trap_no       uint32
	error_code    uint32
	oldmask       uint32
	r0            uint32
	r1            uint32
	r2            uint32
	r3            uint32
	r4            uint32
	r5            uint32
	r6            uint32
	r7            uint32
	r8            uint32
	r9            uint32
	r10           uint32
	fp            uint32
	ip            uint32
	sp            uint32
	lr            uint32
	pc            uint32
	cpsr          uint32
	fault_address uint32
}

type ucontext struct {
	uc_flags    uint32
	uc_link     *ucontext
	uc_stack    stackt
	uc_mcontext sigcontext
	uc_sigmask  uint32
	__unused    [31]int32
	uc_regspace [128]uint32
}

type timeval struct {
	tv_sec  int32
	tv_usec int32
}

func (tv *timeval) set_usec(x int32) {
	tv.tv_usec = x
}

type itimerval struct {
	it_interval timeval
	it_value    timeval
}

type siginfo struct {
	si_signo int32
	si_errno int32
	si_code  int32
	// below here is a union; si_addr is the only field we use
	si_addr uint32
	// si_fd is the field right after the union
	si_fd int32
}

type sigactiont struct {
	sa_handler  uintptr
	sa_flags    uint32
	sa_restorer uintptr
	sa_mask     uint64
}

type epollevent struct {
	events uint32
	_pad   uint32
	data   [8]byte // to match amd64
}

type sockaddr_un struct {
	family uint16
	path   [108]byte
}

type fOwnerEx struct {
	_type int32
	pid   int32
}
