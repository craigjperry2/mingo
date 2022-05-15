package orchestrator

import (
	"sync"
	"sync/atomic"
)

// I'm faking an enum with all this messing around, users can't see it's just an int64
type lifecycle int64

const (
	LifecycleStarting lifecycle = iota + 1
	LifecycleRunning
	LifecycleStopping
)

var instance lifecycle // all access must be atomic for thread safety
var once sync.Once

// GetLifecycleState is called by any code that needs to get access to the application's lifecycle state
// This method is threadsafe even though the underlying field is a singleton
// It returns a defensive copy wrapped in the opaque lifecycle type "enum"
func GetLifecycleState() lifecycle {
	once.Do(func() {
		atomic.StoreInt64((*int64)(&instance), int64(LifecycleStarting))
	})
	return lifecycle(atomic.LoadInt64((*int64)(&instance)))
}

// attemptTransitionToRunning is called by the app orchestrator to change the application lifecycle state
// Valid transitions:
//	* Starting -> Running
// This method is threadsafe, the app becomes mutli-threaded during bootstrapping once the signal handler is created
// It returns false when the requested state transition was rejected, otherwise returns true
func attemptTransitionToRunning() bool {
	return atomic.CompareAndSwapInt64((*int64)(&instance), int64(LifecycleStarting), int64(LifecycleRunning))
}

// transitionToStopping is called by the app orchestrator to change the application lifecycle state
// Valid state transitions:
//	* From Starting -> Stopping (e.g. CTRL+C while bootstrapping)
//	* From Running -> Stopping (e.g. CTRL+C while running)
// This method is threadsafe to allow for clients concurrently calling Get() during Running state
func transitionToStopping() {
	atomic.SwapInt64((*int64)(&instance), int64(LifecycleStopping))
}
