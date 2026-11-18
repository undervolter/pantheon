//go:build windows

package lock

import (
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"
)

const errorAlreadyExists = 183

// lockFile uses a named mutex: it needs no file on disk and disappears with
// the process.
func lockFile(path string) (func(), error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		abs = path
	}
	name, err := syscall.UTF16PtrFromString("Local\\pantheon-" + strings.NewReplacer("\\", "_", ":", "_", "/", "_").Replace(strings.ToLower(abs)))
	if err != nil {
		return nil, err
	}
	proc := syscall.NewLazyDLL("kernel32.dll").NewProc("CreateMutexW")
	handle, _, callErr := proc.Call(0, 0, uintptr(unsafe.Pointer(name)))
	if handle == 0 {
		return nil, callErr
	}
	if errno, ok := callErr.(syscall.Errno); ok && errno == errorAlreadyExists {
		_ = syscall.CloseHandle(syscall.Handle(handle))
		return nil, ErrLocked
	}
	return func() { _ = syscall.CloseHandle(syscall.Handle(handle)) }, nil
}
