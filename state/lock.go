package state

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/shapedthought/vcli/utils"
)

const (
	lockTimeout = 5 * time.Minute
	lockFile    = ".vcli/state.lock"
)

// Lock represents a state lock
type Lock struct {
	lockPath string
	acquired bool
}

// NewLock creates a new lock instance
func NewLock() *Lock {
	settingsPath := utils.SettingPath()
	return &Lock{
		lockPath: filepath.Join(settingsPath, lockFile),
		acquired: false,
	}
}

// Acquire attempts to acquire the lock
func (l *Lock) Acquire() error {
	// Ensure .vcli directory exists
	dir := filepath.Dir(l.lockPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create state directory: %w", err)
	}

	// Check if lock file exists
	if info, err := os.Stat(l.lockPath); err == nil {
		// Lock file exists, check if it's stale
		if time.Since(info.ModTime()) > lockTimeout {
			// Stale lock, remove it
			if err := os.Remove(l.lockPath); err != nil {
				return fmt.Errorf("failed to remove stale lock: %w", err)
			}
		} else {
			// Lock is held by another process
			return fmt.Errorf("state is locked by another process (lock file: %s)", l.lockPath)
		}
	}

	// Create lock file
	file, err := os.Create(l.lockPath)
	if err != nil {
		return fmt.Errorf("failed to create lock file: %w", err)
	}
	defer file.Close()

	// Write current timestamp
	if _, err := file.WriteString(time.Now().Format(time.RFC3339)); err != nil {
		return fmt.Errorf("failed to write to lock file: %w", err)
	}

	l.acquired = true
	return nil
}

// Release releases the lock
func (l *Lock) Release() error {
	if !l.acquired {
		return nil
	}

	if err := os.Remove(l.lockPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove lock file: %w", err)
	}

	l.acquired = false
	return nil
}

// IsAcquired returns whether the lock is currently acquired
func (l *Lock) IsAcquired() bool {
	return l.acquired
}
