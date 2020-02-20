// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pprof

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"runtime"
	"runtime/pprof"
	"strings"
	"testing"
)

// TestDescriptions checks that the profile names under runtime/pprof package
// have a key in the description map.
func TestDescriptions(t *testing.T) {
	for _, p := range pprof.Profiles() {
		_, ok := profileDescriptions[p.Name()]
		if ok != true {
			t.Errorf("%s does not exist in profileDescriptions map\n", p.Name())
		}
	}
}

func isRunnablePlatform() bool {
	if runtime.GOOS != "linux" {
		return false
	}
	if runtime.GOARCH != "amd64" && runtime.GOARCH != "arm64" && runtime.GOARCH != "386" {
		return false
	}
	// Is it in QEMU?
	// Ignore if the tests are running in a QEMU-based emulator,
	// IN_QEMU environmental variable is set by some of the Go builders.
	// IN_QEMU=1 indicates that the tests are running in QEMU.
	if os.Getenv("IN_QEMU") == "1" {
		return false
	}

	//is it in a VM?
	out, err := exec.Command("lscpu").CombinedOutput() // Linux only
	if err != nil {
		return false
	}
	return !strings.Contains(string(out), "Hypervisor")
}

func TestHandlers(t *testing.T) {
	testCases := []struct {
		checkPlatform      bool
		path               string
		handler            http.HandlerFunc
		statusCode         int
		contentType        string
		contentDisposition string
		resp               []byte
	}{
		{false, "/debug/pprof/<script>scripty<script>", Index, http.StatusNotFound, "text/plain; charset=utf-8", "", []byte("Unknown profile\n")},
		{false, "/debug/pprof/heap", Index, http.StatusOK, "application/octet-stream", `attachment; filename="heap"`, nil},
		{false, "/debug/pprof/heap?debug=1", Index, http.StatusOK, "text/plain; charset=utf-8", "", nil},
		{false, "/debug/pprof/cmdline", Cmdline, http.StatusOK, "text/plain; charset=utf-8", "", nil},
		{false, "/debug/pprof/profile?seconds=1", Profile, http.StatusOK, "application/octet-stream", `attachment; filename="profile"`, nil},
		{false, "/debug/pprof/symbol", Symbol, http.StatusOK, "text/plain; charset=utf-8", "", nil},
		{false, "/debug/pprof/trace", Trace, http.StatusOK, "application/octet-stream", `attachment; filename="trace"`, nil},
		{true, "/debug/pprof/profile?seconds=1&event=timer", Profile, http.StatusOK, "application/octet-stream", `attachment; filename="profile"`, nil},
		{true, "/debug/pprof/profile?seconds=1&event=cycles&period=100000", Profile, http.StatusOK, "application/octet-stream", `attachment; filename="profile"`, nil},
	}
	for _, tc := range testCases {
		t.Run(tc.path, func(t *testing.T) {
			if tc.checkPlatform && !isRunnablePlatform() {
				// Skip
			} else {
				req := httptest.NewRequest("GET", "http://example.com"+tc.path, nil)
				w := httptest.NewRecorder()
				tc.handler(w, req)

				resp := w.Result()
				if got, want := resp.StatusCode, tc.statusCode; got != want {
					t.Errorf("status code: got %d; want %d", got, want)
				}

				body, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					t.Errorf("when reading response body, expected non-nil err; got %v", err)
				}
				if got, want := resp.Header.Get("X-Content-Type-Options"), "nosniff"; got != want {
					t.Errorf("X-Content-Type-Options: got %q; want %q", got, want)
				}
				if got, want := resp.Header.Get("Content-Type"), tc.contentType; got != want {
					t.Errorf("Content-Type: got %q; want %q", got, want)
				}
				if got, want := resp.Header.Get("Content-Disposition"), tc.contentDisposition; got != want {
					t.Errorf("Content-Disposition: got %q; want %q", got, want)
				}

				if resp.StatusCode == http.StatusOK {
					return
				}
				if got, want := resp.Header.Get("X-Go-Pprof"), "1"; got != want {
					t.Errorf("X-Go-Pprof: got %q; want %q", got, want)
				}
				if !bytes.Equal(body, tc.resp) {
					t.Errorf("response: got %q; want %q", body, tc.resp)
				}
			}
		})
	}

}
