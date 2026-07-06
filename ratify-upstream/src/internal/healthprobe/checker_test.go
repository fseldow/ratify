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

package healthprobe

import (
	"errors"
	"testing"
)

func TestNewChecker_Valid(t *testing.T) {
	checker, err := NewChecker("test-checker", func() error {
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if checker == nil {
		t.Fatal("expected non-nil checker")
	}
	if checker.Name() != "test-checker" {
		t.Errorf("expected name 'test-checker', got %q", checker.Name())
	}
}

func TestNewChecker_EmptyName(t *testing.T) {
	_, err := NewChecker("", func() error { return nil })
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestNewChecker_NilFunc(t *testing.T) {
	_, err := NewChecker("test", nil)
	if err == nil {
		t.Fatal("expected error for nil function")
	}
}

func TestMustNewChecker_Valid(t *testing.T) {
	checker := MustNewChecker("must-checker", func() error { return nil })
	if checker == nil {
		t.Fatal("expected non-nil checker")
	}
	if checker.Name() != "must-checker" {
		t.Errorf("expected name 'must-checker', got %q", checker.Name())
	}
}

func TestMustNewChecker_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for invalid checker")
		}
	}()
	MustNewChecker("", func() error { return nil })
}

func TestCheckerFunc_Check_Success(t *testing.T) {
	checker, _ := NewChecker("ok", func() error { return nil })
	if err := checker.Check(); err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
}

func TestCheckerFunc_Check_Error(t *testing.T) {
	checker, _ := NewChecker("failing", func() error {
		return errors.New("something went wrong")
	})
	if err := checker.Check(); err == nil {
		t.Fatal("expected error from failing checker")
	}
}

func TestCheckerFunc_NilReceiver(t *testing.T) {
	var checker *CheckerFunc
	if name := checker.Name(); name != "" {
		t.Errorf("expected empty name for nil checker, got %q", name)
	}
	if err := checker.Check(); err == nil {
		t.Fatal("expected error for nil checker Check()")
	}
}

func TestRegistry_RegisterLiveness(t *testing.T) {
	reg := NewRegistry()
	checker := MustNewChecker("live-1", func() error { return nil })

	if err := reg.RegisterLiveness(checker); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	checkers := reg.LivenessCheckers()
	if len(checkers) != 1 {
		t.Fatalf("expected 1 liveness checker, got %d", len(checkers))
	}
	if checkers[0].Name() != "live-1" {
		t.Errorf("expected checker name 'live-1', got %q", checkers[0].Name())
	}
}

func TestRegistry_RegisterReadiness(t *testing.T) {
	reg := NewRegistry()
	checker := MustNewChecker("ready-1", func() error { return nil })

	if err := reg.RegisterReadiness(checker); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	checkers := reg.ReadinessCheckers()
	if len(checkers) != 1 {
		t.Fatalf("expected 1 readiness checker, got %d", len(checkers))
	}
	if checkers[0].Name() != "ready-1" {
		t.Errorf("expected checker name 'ready-1', got %q", checkers[0].Name())
	}
}

func TestRegistry_DuplicateRegistration(t *testing.T) {
	reg := NewRegistry()
	checker := MustNewChecker("dup", func() error { return nil })

	if err := reg.RegisterLiveness(checker); err != nil {
		t.Fatalf("first registration should succeed: %v", err)
	}
	if err := reg.RegisterLiveness(checker); err == nil {
		t.Fatal("expected error for duplicate registration")
	}
}

func TestRegistry_RegisterNilChecker(t *testing.T) {
	reg := NewRegistry()
	if err := reg.RegisterLiveness(nil); err == nil {
		t.Fatal("expected error for nil checker")
	}
}

func TestRegistry_NilReceiver(t *testing.T) {
	var reg *Registry
	if err := reg.RegisterLiveness(MustNewChecker("x", func() error { return nil })); err == nil {
		t.Fatal("expected error for nil registry")
	}
	if checkers := reg.LivenessCheckers(); checkers != nil {
		t.Fatalf("expected nil for nil registry LivenessCheckers, got %v", checkers)
	}
	if checkers := reg.ReadinessCheckers(); checkers != nil {
		t.Fatalf("expected nil for nil registry ReadinessCheckers, got %v", checkers)
	}
}

func TestRegistry_MultipleCheckers(t *testing.T) {
	reg := NewRegistry()
	for i := 0; i < 5; i++ {
		name := "checker-" + string(rune('a'+i))
		checker := MustNewChecker(name, func() error { return nil })
		if err := reg.RegisterReadiness(checker); err != nil {
			t.Fatalf("failed to register checker %s: %v", name, err)
		}
	}

	checkers := reg.ReadinessCheckers()
	if len(checkers) != 5 {
		t.Fatalf("expected 5 readiness checkers, got %d", len(checkers))
	}
}

func TestRegistry_SnapshotIsolation(t *testing.T) {
	reg := NewRegistry()
	checker := MustNewChecker("iso", func() error { return nil })
	if err := reg.RegisterLiveness(checker); err != nil {
		t.Fatal(err)
	}

	// Get a snapshot
	snapshot := reg.LivenessCheckers()

	// Add another checker
	checker2 := MustNewChecker("iso-2", func() error { return nil })
	if err := reg.RegisterLiveness(checker2); err != nil {
		t.Fatal(err)
	}

	// Original snapshot should not be affected
	if len(snapshot) != 1 {
		t.Fatalf("snapshot should be isolated, expected 1, got %d", len(snapshot))
	}
}
