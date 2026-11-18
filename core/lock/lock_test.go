package lock

import (
	"errors"
	"path/filepath"
	"testing"
)

func TestSecondAcquireIsRefused(t *testing.T) {
	path := filepath.Join(t.TempDir(), "pantheon.lock")

	release, err := Acquire(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Acquire(path); !errors.Is(err, ErrLocked) {
		t.Fatalf("second acquire: err = %v, want ErrLocked", err)
	}

	release()
	again, err := Acquire(path)
	if err != nil {
		t.Fatalf("the lock must be free after release: %v", err)
	}
	again()
}

func TestDifferentDirectoriesDoNotCollide(t *testing.T) {
	a, err := Acquire(filepath.Join(t.TempDir(), "pantheon.lock"))
	if err != nil {
		t.Fatal(err)
	}
	defer a()
	b, err := Acquire(filepath.Join(t.TempDir(), "pantheon.lock"))
	if err != nil {
		t.Fatalf("another data directory must have its own lock: %v", err)
	}
	b()
}
