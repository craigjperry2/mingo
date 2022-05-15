package orchestrator

import (
	"sync/atomic"
	"testing"
)

func TestRegularLifecycleTransitions(t *testing.T) {
	teardown := setupLifecycle()
	defer teardown()

	got := GetLifecycleState()

	// Should begin in Starting state
	if got != LifecycleStarting {
		t.Errorf("starting want %d, got %d", LifecycleStarting, got)
	}

	// Should transition successfully to Running state
	success := attemptTransitionToRunning()
	if !success {
		t.Error("unexpected failure in transition to running")
	}
	got = GetLifecycleState()
	if got != LifecycleRunning {
		t.Errorf("running want %d, got %d", LifecycleRunning, got)
	}

	// Should transition to Stopping state
	transitionToStopping()
	got = GetLifecycleState()
	if got != LifecycleStopping {
		t.Errorf("stopping want %d, got %d", LifecycleStopping, got)
	}

	// Transition to Stopping state should be idempotent
	transitionToStopping()
	got = GetLifecycleState()
	if got != LifecycleStopping {
		t.Errorf("idempotent want %d, got %d", LifecycleStopping, got)
	}

	// Transition backward to Running state should not succeed
	success = attemptTransitionToRunning()
	if success {
		t.Error("invalid state transition from stopping to running")
	}
}

func TestStartingToStoppingLifecycleTransitions(t *testing.T) {
	teardown := setupLifecycle()
	defer teardown()

	got := GetLifecycleState()

	// Should begin in Starting state
	if got != LifecycleStarting {
		t.Errorf("starting want %d, got %d", LifecycleStarting, got)
	}

	// Should transition to Stopping state
	transitionToStopping()
	got = GetLifecycleState()
	if got != LifecycleStopping {
		t.Errorf("stopping want %d, got %d", LifecycleStopping, got)
	}

	// Transition to Stopping state should be idempotent
	transitionToStopping()
	got = GetLifecycleState()
	if got != LifecycleStopping {
		t.Errorf("idempotent want %d, got %d", LifecycleStopping, got)
	}

	// Transition backward to Running state should not succeed
	success := attemptTransitionToRunning()
	if success {
		t.Error("invalid state transition from stopping to running")
	}
}

func setupLifecycle() func() {
	// noop to catch flaky tests that don't cleanup after themselves

	return func() {
		atomic.StoreInt64((*int64)(&instance), int64(LifecycleStarting))
	}
}
