package orchestrator

import (
	"io/ioutil"
	"syscall"
	"testing"
	"time"
)

func TestSignalHandlerMovesLifecycleToStopping(t *testing.T) {
	l := GetLifecycleState()
	if l != LifecycleStarting {
		t.Errorf("start want LifecycleStarting, got %v", l)
	}

	go Orchestrate([]string{}, ioutil.Discard)
	time.Sleep(100 * time.Millisecond) // await setup

	syscall.Kill(syscall.Getpid(), syscall.SIGINT)
	time.Sleep(100 * time.Millisecond) // await delivery

	l = GetLifecycleState()
	if l != LifecycleStopping {
		t.Errorf("stop want LifecycleStopping, got %v", l)
	}
}
