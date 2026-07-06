/*
Copyright The Ratify Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package manager

import (
	"testing"
	"time"
)

func TestNewReadySignal(t *testing.T) {
	rs := NewReadySignal()
	if rs == nil {
		t.Fatal("expected non-nil ReadySignal")
	}
	if rs.IsReady() {
		t.Error("expected not ready initially")
	}
}

func TestReadySignal_MarkReady(t *testing.T) {
	rs := NewReadySignal()
	rs.MarkReady()

	if !rs.IsReady() {
		t.Error("expected ready after MarkReady()")
	}
}

func TestReadySignal_MarkReady_Idempotent(t *testing.T) {
	rs := NewReadySignal()
	rs.MarkReady()
	rs.MarkReady() // should not panic (double close)

	if !rs.IsReady() {
		t.Error("expected ready after multiple MarkReady() calls")
	}
}

func TestReadySignal_Done_ClosesOnReady(t *testing.T) {
	rs := NewReadySignal()
	done := rs.Done()

	// Channel should not be closed yet
	select {
	case <-done:
		t.Fatal("Done() channel should not be closed before MarkReady()")
	default:
	}

	rs.MarkReady()

	// Channel should now be closed
	select {
	case <-done:
		// expected
	case <-time.After(1 * time.Second):
		t.Fatal("Done() channel should be closed after MarkReady()")
	}
}

func TestReadySignal_NilReceiver_Done(t *testing.T) {
	var rs *ReadySignal
	if ch := rs.Done(); ch != nil {
		t.Error("expected nil channel for nil receiver")
	}
}

func TestReadySignal_NilReceiver_MarkReady(t *testing.T) {
	var rs *ReadySignal
	// Should not panic
	rs.MarkReady()
}

func TestReadySignal_NilReceiver_IsReady(t *testing.T) {
	var rs *ReadySignal
	if rs.IsReady() {
		t.Error("expected false for nil receiver")
	}
}

func TestReadySignal_Checker_NotReady(t *testing.T) {
	rs := NewReadySignal()
	checker := rs.Checker()

	if checker == nil {
		t.Fatal("expected non-nil checker")
	}
	if checker.Name() != "controller-runtime-manager" {
		t.Errorf("expected name 'controller-runtime-manager', got %q", checker.Name())
	}

	if err := checker.Check(); err == nil {
		t.Error("expected error from checker when not ready")
	}
}

func TestReadySignal_Checker_Ready(t *testing.T) {
	rs := NewReadySignal()
	rs.MarkReady()

	checker := rs.Checker()
	if err := checker.Check(); err != nil {
		t.Errorf("expected nil error from checker when ready, got: %v", err)
	}
}

func TestReadySignal_Checker_NilReceiver(t *testing.T) {
	var rs *ReadySignal
	checker := rs.Checker()

	if err := checker.Check(); err == nil {
		t.Error("expected error from checker with nil receiver")
	}
}

func TestReadySignal_ConcurrentAccess(t *testing.T) {
	rs := NewReadySignal()
	done := make(chan struct{})

	// Concurrent readers
	for i := 0; i < 10; i++ {
		go func() {
			for {
				select {
				case <-done:
					return
				default:
					rs.IsReady()
				}
			}
		}()
	}

	// Writer
	go func() {
		time.Sleep(10 * time.Millisecond)
		rs.MarkReady()
	}()

	// Wait for ready
	select {
	case <-rs.Done():
		// success
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for ready signal")
	}

	close(done)
}
