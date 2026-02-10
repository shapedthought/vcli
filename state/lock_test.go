package state

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// setupLockTest creates a temp dir and sets OWLCTL_SETTINGS_PATH.
// Returns the temp dir path and a cleanup function.
func setupLockTest(t *testing.T) (string, func()) {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "owlctl-lock-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	origEnv := os.Getenv("OWLCTL_SETTINGS_PATH")
	os.Setenv("OWLCTL_SETTINGS_PATH", tmpDir)

	cleanup := func() {
		os.Setenv("OWLCTL_SETTINGS_PATH", origEnv)
		os.RemoveAll(tmpDir)
	}
	return tmpDir, cleanup
}

func TestLockAcquireAndRelease(t *testing.T) {
	_, cleanup := setupLockTest(t)
	defer cleanup()

	l := NewLock()

	if err := l.Acquire(); err != nil {
		t.Fatalf("Acquire failed: %v", err)
	}
	if !l.IsAcquired() {
		t.Error("Expected IsAcquired=true after Acquire")
	}

	if err := l.Release(); err != nil {
		t.Fatalf("Release failed: %v", err)
	}
	if l.IsAcquired() {
		t.Error("Expected IsAcquired=false after Release")
	}
}

func TestLockDoubleAcquireFails(t *testing.T) {
	_, cleanup := setupLockTest(t)
	defer cleanup()

	l1 := NewLock()
	if err := l1.Acquire(); err != nil {
		t.Fatalf("First Acquire failed: %v", err)
	}
	defer l1.Release()

	l2 := NewLock()
	err := l2.Acquire()
	if err == nil {
		t.Error("Expected error on double acquire")
		l2.Release()
	}
}

func TestLockReleaseWithoutAcquire(t *testing.T) {
	_, cleanup := setupLockTest(t)
	defer cleanup()

	l := NewLock()

	// Should not error (idempotent)
	if err := l.Release(); err != nil {
		t.Errorf("Expected no error on release without acquire, got: %v", err)
	}
}

func TestLockStaleLockCleanup(t *testing.T) {
	tmpDir, cleanup := setupLockTest(t)
	defer cleanup()

	// Create the lock file directory and a stale lock file
	lockDir := filepath.Join(tmpDir, ".owlctl")
	if err := os.MkdirAll(lockDir, 0755); err != nil {
		t.Fatalf("Failed to create lock dir: %v", err)
	}

	lockPath := filepath.Join(lockDir, "state.lock")
	if err := os.WriteFile(lockPath, []byte("stale"), 0644); err != nil {
		t.Fatalf("Failed to create stale lock file: %v", err)
	}

	// Set mtime to 10 minutes ago (beyond the 5-minute lockTimeout)
	staleTime := time.Now().Add(-10 * time.Minute)
	if err := os.Chtimes(lockPath, staleTime, staleTime); err != nil {
		t.Fatalf("Failed to set stale mtime: %v", err)
	}

	l := NewLock()
	if err := l.Acquire(); err != nil {
		t.Fatalf("Expected Acquire to succeed with stale lock, got: %v", err)
	}
	defer l.Release()

	if !l.IsAcquired() {
		t.Error("Expected IsAcquired=true after acquiring stale lock")
	}
}

func TestLockIsAcquiredLifecycle(t *testing.T) {
	_, cleanup := setupLockTest(t)
	defer cleanup()

	l := NewLock()

	if l.IsAcquired() {
		t.Error("Expected IsAcquired=false initially")
	}

	l.Acquire()
	if !l.IsAcquired() {
		t.Error("Expected IsAcquired=true after Acquire")
	}

	l.Release()
	if l.IsAcquired() {
		t.Error("Expected IsAcquired=false after Release")
	}
}
