// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

const (
	_SS_DISABLE  = 4
	_NSIG        = 33
	_SI_USER     = 0x10001
	_POLL_IN     = 0x1 // taken from https://github.com/freebsd/freebsd/blob/2e1c48e4b2db19ac271c688a4145fd41348f0374/sys/sys/signal.h
	_SIG_BLOCK   = 1
	_SIG_UNBLOCK = 2
	_SIG_SETMASK = 3
)
