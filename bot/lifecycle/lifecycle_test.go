package lifecycle

import (
	"testing"
)

func TestLifecycleStartStopUpdate(t *testing.T) {
	lc := New()

	// Mock the running state functions
	isRunning := false
	lc.SetRunningFunc(func() bool {
		return isRunning
	}, func(status bool) {
		isRunning = status
	})

	// Create a goroutine to listen to the startSignal to prevent blocking
	go func() {
		<-lc.StartSignal()
	}()

	// Test Start
	lc.Start()
	if !isRunning {
		t.Error("Expected lifecycle to be set to running")
	}

	// Create a goroutine to listen to the stopSignal to prevent blocking
	go func() {
		<-lc.StopSignal()
	}()

	// Test Stop
	lc.Stop()
	if isRunning {
		t.Error("Expected lifecycle to be stopped")
	}

	// Update doesn't change the isRunning state, but we listen to prevent blocking
	go func() {
		<-lc.ReloadSignal()
	}()

	// Test Update
	lc.Update()
}

func TestLifecycleSingleton(t *testing.T) {
	lc1 := New()
	lc2 := New()

	if lc1 != lc2 {
		t.Error("Expected New() to return the same instance of Client")
	}
}
