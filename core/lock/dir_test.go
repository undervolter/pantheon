package lock

import (
	"os"
	"path/filepath"
	"testing"
)

// The lock is taken before the database is opened, so it is the first thing
// to need the data directory. On Unix a missing directory made the very
// first start with a custom data_dir fail.
func TestAcquireCreatesTheDataDirectory(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "data", "nested")
	release, err := Acquire(filepath.Join(dir, "pantheon.lock"))
	if err != nil {
		t.Fatalf("acquire in a directory that does not exist yet: %v", err)
	}
	defer release()
	if info, err := os.Stat(dir); err != nil || !info.IsDir() {
		t.Fatalf("the data directory was not created: %v", err)
	}
}
