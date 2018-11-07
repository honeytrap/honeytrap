// Copyright 2018 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// +build linux,!amd64

package rawfile

import (
	"syscall"
	"unsafe"
)

func blockingPoll(fds *pollEvent, nfds int, timeout int64) (int, syscall.Errno) {
	var ts *syscall.Timespec = nil

	if timeout != -1 {
		timeSpec := syscall.NsecToTimespec(timeout * 1000000)
		ts = &timeSpec
	}

	// we are using SYS_PPOLL here instead of SYS_POLL, because SYS_POLL isn't available on ARM64
	n, _, e := syscall.Syscall6(syscall.SYS_PPOLL, uintptr(unsafe.Pointer(fds)), uintptr(nfds), uintptr(unsafe.Pointer(ts)), 0, 0, 0)
	return int(n), e
}
