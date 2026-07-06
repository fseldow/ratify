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
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewServer_Valid(t *testing.T) {
	reg := NewRegistry()
	srv, err := NewServer(":9090", reg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if srv == nil {
		t.Fatal("expected non-nil server")
	}
}

func TestNewServer_EmptyAddress(t *testing.T) {
	_, err := NewServer("", NewRegistry())
	if err == nil {
		t.Fatal("expected error for empty address")
	}
}

func TestNewServer_NilRegistry(t *testing.T) {
	srv, err := NewServer(":9090", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if srv == nil {
		t.Fatal("expected non-nil server with default registry")
	}
}

func TestServer_HealthzHandler_AllHealthy(t *testing.T) {
	reg := NewRegistry()
	checker := MustNewChecker("test-live", func() error { return nil })
	if err := reg.RegisterLiveness(checker); err != nil {
		t.Fatal(err)
	}

	srv, err := NewServer(":0", reg)
	if err != nil {
		t.Fatal(err)
	}
	srv.started.Store(true)

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()
	srv.mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var resp response
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Status != "ok" {
		t.Errorf("expected status 'ok', got %q", resp.Status)
	}
	if len(resp.Checks) != 1 {
		t.Fatalf("expected 1 check result, got %d", len(resp.Checks))
	}
	if resp.Checks[0].Name != "test-live" {
		t.Errorf("expected check name 'test-live', got %q", resp.Checks[0].Name)
	}
}

func TestServer_HealthzHandler_Unhealthy(t *testing.T) {
	reg := NewRegistry()
	checker := MustNewChecker("failing", func() error {
		return errors.New("component is down")
	})
	if err := reg.RegisterLiveness(checker); err != nil {
		t.Fatal(err)
	}

	srv, err := NewServer(":0", reg)
	if err != nil {
		t.Fatal(err)
	}
	srv.started.Store(true)

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()
	srv.mux.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503, got %d", w.Code)
	}

	var resp response
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Status != "not alive" {
		t.Errorf("expected status 'not alive', got %q", resp.Status)
	}
	if len(resp.Checks) == 0 {
		t.Fatal("expected check results in response")
	}
	if resp.Checks[0].Error != "component is down" {
		t.Errorf("expected error 'component is down', got %q", resp.Checks[0].Error)
	}
}

func TestServer_HealthzHandler_NotStarted(t *testing.T) {
	reg := NewRegistry()
	srv, err := NewServer(":0", reg)
	if err != nil {
		t.Fatal(err)
	}
	// Don't mark as started

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()
	srv.mux.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503 before start, got %d", w.Code)
	}
}

func TestServer_ReadyzHandler_Ready(t *testing.T) {
	reg := NewRegistry()
	checker := MustNewChecker("executor", func() error { return nil })
	if err := reg.RegisterReadiness(checker); err != nil {
		t.Fatal(err)
	}

	srv, err := NewServer(":0", reg)
	if err != nil {
		t.Fatal(err)
	}
	srv.started.Store(true)

	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	w := httptest.NewRecorder()
	srv.mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var resp response
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Status != "ok" {
		t.Errorf("expected status 'ok', got %q", resp.Status)
	}
}

func TestServer_ReadyzHandler_NoCheckers(t *testing.T) {
	reg := NewRegistry()
	srv, err := NewServer(":0", reg)
	if err != nil {
		t.Fatal(err)
	}
	srv.started.Store(true)

	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	w := httptest.NewRecorder()
	srv.mux.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503 with no readiness checkers, got %d", w.Code)
	}

	var resp response
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Status != "not ready" {
		t.Errorf("expected status 'not ready', got %q", resp.Status)
	}
}

func TestServer_ReadyzHandler_CheckerFails(t *testing.T) {
	reg := NewRegistry()
	checker := MustNewChecker("executor", func() error {
		return errors.New("executor not loaded")
	})
	if err := reg.RegisterReadiness(checker); err != nil {
		t.Fatal(err)
	}

	srv, err := NewServer(":0", reg)
	if err != nil {
		t.Fatal(err)
	}
	srv.started.Store(true)

	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	w := httptest.NewRecorder()
	srv.mux.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503, got %d", w.Code)
	}

	var resp response
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Status != "not ready" {
		t.Errorf("expected status 'not ready', got %q", resp.Status)
	}
}

func TestServer_ReadyzHandler_NotStarted(t *testing.T) {
	reg := NewRegistry()
	checker := MustNewChecker("executor", func() error { return nil })
	if err := reg.RegisterReadiness(checker); err != nil {
		t.Fatal(err)
	}

	srv, err := NewServer(":0", reg)
	if err != nil {
		t.Fatal(err)
	}
	// Not started

	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	w := httptest.NewRecorder()
	srv.mux.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503, got %d", w.Code)
	}
}

func TestServer_Run_ContextCanceled(t *testing.T) {
	reg := NewRegistry()
	srv, err := NewServer(":0", reg)
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.Run(ctx)
	}()

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("expected nil error on context cancel, got: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("server did not shut down within timeout")
	}
}

func TestServer_Run_NilServer(t *testing.T) {
	var srv *Server
	err := srv.Run(context.Background())
	if err == nil {
		t.Fatal("expected error for nil server")
	}
}

func TestServer_Run_NilContext(t *testing.T) {
	reg := NewRegistry()
	srv, err := NewServer(":0", reg)
	if err != nil {
		t.Fatal(err)
	}

	err = srv.Run(nil)
	if err == nil {
		t.Fatal("expected error for nil context")
	}
}

func TestEvaluate_NilChecker(t *testing.T) {
	checkers := []HealthChecker{nil}
	results, healthy := evaluate(checkers)

	if healthy {
		t.Error("expected unhealthy when nil checker present")
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Name != "unknown" {
		t.Errorf("expected name 'unknown', got %q", results[0].Name)
	}
	if results[0].Status != "error" {
		t.Errorf("expected status 'error', got %q", results[0].Status)
	}
}

func TestEvaluate_MixedCheckers(t *testing.T) {
	checkers := []HealthChecker{
		MustNewChecker("healthy", func() error { return nil }),
		MustNewChecker("failing", func() error { return errors.New("oops") }),
	}

	results, healthy := evaluate(checkers)
	if healthy {
		t.Error("expected unhealthy when any checker fails")
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if results[0].Status != "ok" {
		t.Errorf("expected first check 'ok', got %q", results[0].Status)
	}
	if results[1].Status != "error" {
		t.Errorf("expected second check 'error', got %q", results[1].Status)
	}
}

func TestEvaluate_EmptyCheckers(t *testing.T) {
	results, healthy := evaluate([]HealthChecker{})
	if !healthy {
		t.Error("expected healthy when no checkers registered")
	}
	if len(results) != 0 {
		t.Fatalf("expected 0 results, got %d", len(results))
	}
}

func TestServer_HealthzHandler_NoLivenessCheckers(t *testing.T) {
	reg := NewRegistry()
	srv, err := NewServer(":0", reg)
	if err != nil {
		t.Fatal(err)
	}
	srv.started.Store(true)

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()
	srv.mux.ServeHTTP(w, req)

	// With no liveness checkers, evaluate returns healthy=true
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 with no liveness checkers, got %d", w.Code)
	}
}

func TestServer_HealthzHandler_MultipleLivenessCheckers(t *testing.T) {
	reg := NewRegistry()
	if err := reg.RegisterLiveness(MustNewChecker("check-a", func() error { return nil })); err != nil {
		t.Fatal(err)
	}
	if err := reg.RegisterLiveness(MustNewChecker("check-b", func() error { return nil })); err != nil {
		t.Fatal(err)
	}

	srv, err := NewServer(":0", reg)
	if err != nil {
		t.Fatal(err)
	}
	srv.started.Store(true)

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()
	srv.mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var resp response
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if len(resp.Checks) != 2 {
		t.Errorf("expected 2 checks, got %d", len(resp.Checks))
	}
}
