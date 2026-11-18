// Package lock keeps a second copy of the bot from starting on the same data
// directory. Two copies would both poll Telegram (it answers 409 to one of
// them) and both write the same database.
package lock

import "errors"

// ErrLocked means another process already holds the lock.
var ErrLocked = errors.New("another pantheon instance is already running on this data directory")

// Acquire takes the lock at path. The returned function releases it; the
// operating system also releases it when the process dies, so a crash never
// leaves a stale lock behind.
func Acquire(path string) (release func(), err error) {
	return lockFile(path)
}
