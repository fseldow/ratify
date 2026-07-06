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

package httpserver

import (
	"testing"

	"github.com/notaryproject/ratify/v2/internal/executor"
)

func TestNewHealthStatus(t *testing.T) {
	hs := NewHealthStatus()
	if hs == nil {
		t.Fatal("expected non-nil HealthStatus")
	}
	if hs.IsAlive() {
		t.Error("expected not alive initially")
	}
}

func TestHealthStatus_MarkAlive(t *testing.T) {
	hs := NewHealthStatus()
	hs.MarkAlive()
	if !hs.IsAlive() {
		t.Error("expected alive after MarkAlive()")
	}
}

func TestHealthStatus_NilReceiver_MarkAlive(t *testing.T) {
	var hs *HealthStatus
	// Should not panic
	hs.MarkAlive()
}

func TestHealthStatus_NilReceiver_IsAlive(t *testing.T) {
	var hs *HealthStatus
	if hs.IsAlive() {
		t.Error("expected false for nil receiver")
	}
}

func TestHealthStatus_AliveChecker_NotAlive(t *testing.T) {
	hs := NewHealthStatus()
	checker := hs.AliveChecker()

	if checker == nil {
		t.Fatal("expected non-nil checker")
	}
	if checker.Name() != httpServerAliveCheckerName {
		t.Errorf("expected name %q, got %q", httpServerAliveCheckerName, checker.Name())
	}
	if err := checker.Check(); err == nil {
		t.Error("expected error when not alive")
	}
}

func TestHealthStatus_AliveChecker_Alive(t *testing.T) {
	hs := NewHealthStatus()
	hs.MarkAlive()
	checker := hs.AliveChecker()

	if err := checker.Check(); err != nil {
		t.Errorf("expected nil error when alive, got: %v", err)
	}
}

func TestHealthStatus_ExecutorChecker_NotAlive(t *testing.T) {
	hs := NewHealthStatus()
	checker := hs.ExecutorChecker(func() *executor.ScopedExecutor {
		return &executor.ScopedExecutor{}
	})

	if err := checker.Check(); err == nil {
		t.Error("expected error when not alive")
	}
}

func TestHealthStatus_ExecutorChecker_NilGetter(t *testing.T) {
	hs := NewHealthStatus()
	hs.MarkAlive()
	checker := hs.ExecutorChecker(nil)

	if err := checker.Check(); err == nil {
		t.Error("expected error for nil executor getter")
	}
}

func TestHealthStatus_ExecutorChecker_NilExecutor(t *testing.T) {
	hs := NewHealthStatus()
	hs.MarkAlive()
	checker := hs.ExecutorChecker(func() *executor.ScopedExecutor {
		return nil
	})

	if err := checker.Check(); err == nil {
		t.Error("expected error when executor is nil")
	}
}

func TestHealthStatus_ExecutorChecker_Valid(t *testing.T) {
	hs := NewHealthStatus()
	hs.MarkAlive()
	checker := hs.ExecutorChecker(func() *executor.ScopedExecutor {
		return &executor.ScopedExecutor{}
	})

	if err := checker.Check(); err != nil {
		t.Errorf("expected nil error with valid executor, got: %v", err)
	}
}

func TestHealthStatus_NilReceiver_AliveChecker(t *testing.T) {
	var hs *HealthStatus
	checker := hs.AliveChecker()

	if err := checker.Check(); err == nil {
		t.Error("expected error for nil HealthStatus")
	}
}

func TestHealthStatus_NilReceiver_ExecutorChecker(t *testing.T) {
	var hs *HealthStatus
	checker := hs.ExecutorChecker(func() *executor.ScopedExecutor {
		return &executor.ScopedExecutor{}
	})

	if err := checker.Check(); err == nil {
		t.Error("expected error for nil HealthStatus")
	}
}
