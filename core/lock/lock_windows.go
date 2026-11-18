//go:build windows

package lock

import (
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"
)

const (
	errorAccessDenied  = 5
	errorAlreadyExists = 183
)

// lockFile uses a named mutex: it needs no file on disk and disappears with
// the process. The name is in the Global namespace so that a copy running as
// a service and one started from a desktop session see each other.
func lockFile(path string) (func(), error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		abs = path
	}
	name, err := syscall.UTF16PtrFromString("Global\\pantheon-" + strings.NewReplacer("\\", "_", ":", "_", "/", "_").Replace(strings.ToLower(abs)))
	if err != nil {
		return nil, err
	}
	proc := syscall.NewLazyDLL("kernel32.dll").NewProc("CreateMutexW")
	handle, _, callErr := proc.Call(0, 0, uintptr(unsafe.Pointer(name)))
	if handle == 0 {
		// The mutex exists but belongs to another account (a service, say),
		// which is exactly the case this lock is for.
		if errno, ok := callErr.(syscall.Errno); ok && errno == errorAccessDenied {
			return nil, ErrLocked
		}
		return nil, callErr
	}
	if errno, ok := callErr.(syscall.Errno); ok && errno == errorAlreadyExists {
		_ = syscall.CloseHandle(syscall.Handle(handle))
		return nil, ErrLocked
	}
	return func() { _ = syscall.CloseHandle(syscall.Handle(handle)) }, nil
}
