# Decisions Log

*Last updated: 2026-06-01T17:00:47+10:00*

---

## Decision: Alpha.2 Deployment Blockers — Design Approach

**Author:** Batty  
**Date:** 2026-06-01  
**Status:** Proposed  
**Scope:** ratify v2.0.0-alpha.2 release readiness

### Context

PR #2524 identified 5 issues during AKS deployment of alpha.1. Four of them are blockers for alpha.2 (health endpoints, CRD packaging, StoreMux routing, cert reload). I've produced a detailed design doc at `ratify-upstream/alpha2-design.md`.

### Key Decisions

1. **Health endpoints at flat paths** (`/healthz`, `/readyz`) — Kubernetes-idiomatic, separate from the `/ratify/gatekeeper/v2/` API namespace.
2. **Helm CRD lifecycle via hook Job** (cert-manager pattern) — not relying on Helm's native `crds/` dir because it never upgrades CRDs.
3. **`"*"` scope as global fallback** — maps to `StoreMux.RegisterFallback()` and a new `ScopedExecutor.fallback` field. Explicit, non-breaking addition.
4. **Trust store cert watcher** — reuse fsnotify pattern from existing `tlssecret/certwatcher.go`, add debounce + periodic poll for K8s symlink-swap mounts.

### Impact

- No breaking changes to existing APIs or configs
- Adds ~5 new files, modifies ~6 existing files
- Unblocks Kubernetes deployment without `--no-hooks` workarounds

### Needs

- Team review of design doc before implementation begins
- Deckard approval for merge to main

---

## Decision: Alpha.2 Blocker Implementation Approach

**Date:** 2026-06-01  
**Author:** Batty  
**Status:** Proposed

### Context

Implemented all 3 deployment blockers from the alpha2-design.md. Key design choices made during implementation:

### Decisions

1. **Health probes use flat paths** (`/healthz`, `/readyz`) — Kubernetes-idiomatic, no auth middleware wrapping. Registered before verify/mutate handlers in gorilla/mux.

2. **`"*"` scope convention** for fallback stores and executors — maps directly to `StoreMux.RegisterFallback()` in ratify-go. Single-store-no-scopes also auto-registers as fallback for zero-config UX.

3. **Cert watcher uses dual detection**: fsnotify for immediate response + SHA-256 polling every 30s for k8s symlink rotation resilience. Debounce window is 2s to coalesce burst writes.

4. **Readiness signal**: polled via goroutine checking `getExecutor() != nil` rather than a channel/callback — simpler, works for both CRD and static config modes without coupling.

### Files Created

- `src/internal/httpserver/health.go` + `health_test.go`
- `src/internal/httpserver/server_modifications.go` (documents server.go changes)
- `src/internal/store/factory_modifications.go` + `factory_test.go`
- `src/internal/executor/executor_modifications.go`
- `src/internal/verifier/truststore/watcher.go` + `watcher_test.go`

---

## Decision: User Directive — CRDs Timing Correction

**Date:** 2026-06-01T16:53:20+10:00  
**By:** xinhl (via Copilot)

### What

CRDs are in Beta milestone, not Alpha.2. Remove Helm CRD Packaging from Alpha.2 scope.

### Why

User correction — CRD lifecycle management is a Beta-tier feature, not an Alpha.2 deployment blocker.

---

## Decision: Ratify v2 Publish — Work Decomposition

**Date:** 2026-06-01  
**Decision by:** Deckard (Lead)  
**Requested by:** xinhl  
**Status:** ACTIVE

### Context

The upstream [Ratify v2 Architecture Proposal](https://github.com/notaryproject/ratify/blob/main/docs/design/Ratify%20v2%20Architecture%20Proposal.md) defines milestones from Alpha.1 through GA. The user requests we "step forward v2 publish" — advance toward the GA release of Ratify v2.

### Current State Assessment

| Milestone | Status | Evidence |
|-----------|--------|----------|
| Alpha.1 — Core library (`ratify-go`) | ✅ Done | Repo exists: executor, Store/Verifier/PolicyEnforcer interfaces, store_mux, oci_store, registry_store |
| Alpha.2 — v2 branch + new Executor in ratify | 🟡 In Progress | `cmd/ratify-gatekeeper-provider` exists; PR #2524 open fixing Dockerfile/Helm |
| Beta.1 — Oras store + Notation verifier | ✅ Done | `ratify-go` has oci/registry stores; `ratify-verifier-go/notation` exists with tests |
| Beta.2 — Store cache + Cosign verifier | 🟡 Partial | `ratify-verifier-go/cosign` exists; store cache status in ratify-go unclear (store_mux.go may serve this) |
| Beta.3 — CLI entrypoint (`ratify-cli`) | ❌ Not started | No `ratify-cli` repo found |
| RC.1 — Missing v1 features | ❌ Not started | Open issues: #2257 (versioned config), #2354 (AWS auth), #2351 (Alibaba auth), #2353 (SBOM), #2352 (vuln report) |
| GA — Final releases | ❌ Not started | No releases tagged on ratify-go; no v2.0.0 on ratify |

### Key Blockers (from PR #2524 deployment testing)

1. No `/healthz` or `/readyz` endpoints — K8s probes kill the pod
2. Helm pre-install hook uses CRD image that doesn't exist for v2
3. StoreMux wildcard `*.` pattern broken — needs explicit registry patterns
4. No Azure auth providers (workload identity, managed identity, K8s secrets)
5. cert-rotator doesn't sync with manually-created ExternalDataProvider caBundle

### Work Items — Grouped by Milestone Phase

#### Phase 1: Complete Alpha.2 (Unblock Deployment)

| Priority | Title | Description | Assigned | Dependencies |
|----------|-------|-------------|----------|--------------|
| P0 | Merge PR #2524 (Dockerfile/Helm fix) | Fix v2 binary name and CLI args in Dockerfile and deployment.yaml | Batty | None |
| P0 | Add health endpoints to v2 provider | Implement `/healthz` and `/readyz` in `cmd/ratify-gatekeeper-provider` to prevent K8s kill loops | Batty | None |
| P0 | Fix Helm CRD installation for v2 | Update pre-install hook or provide v2-compatible CRD manifests that don't depend on v1 CRD image | Batty | None |
| P1 | Fix StoreMux wildcard pattern | Debug and fix `*.` catch-all pattern in store_mux; add proper glob/regex matching for registry patterns | Batty | ratify-go store_mux |
| P1 | cert-rotator + ExternalDataProvider sync | Ensure cert-rotator updates caBundle on the ExternalDataProvider CR after cert regeneration | Batty | None |

#### Phase 2: Complete Beta.2 (Store Cache + Verification)

| Priority | Title | Description | Assigned | Dependencies |
|----------|-------|-------------|----------|--------------|
| P1 | Verify/complete store cache in ratify-go | Confirm store_mux provides caching; if not, implement shared in-memory cache for Oras store that's safe for concurrent verifier access | Batty | ratify-go |
| P1 | Cosign verifier integration test | Write integration tests for cosign verifier in ratify-verifier-go; verify it works end-to-end with K8s provider | Gaff | ratify-verifier-go/cosign |
| P2 | Notation verifier integration test | E2E test for notation verifier through the Gatekeeper provider path | Gaff | ratify-verifier-go/notation |

#### Phase 3: Auth Providers (RC.1 prerequisite)

| Priority | Title | Description | Assigned | Dependencies |
|----------|-------|-------------|----------|--------------|
| P0 | Azure auth provider for v2 | Implement workload identity + managed identity auth in v2 Gatekeeper provider (issue #2354 pattern) | Rachael | Phase 1 complete |
| P1 | AWS ECR auth provider for v2 | Implement AWS ECR auth store (issue #2354) | Rachael | Auth provider interface |
| P2 | Alibaba Cloud RRSA auth provider | Implement Alibaba auth (issue #2351) | Rachael | Auth provider interface |
| P1 | Auth provider security review | Review all auth implementations for credential leakage, token handling, secret rotation | Rachael | Auth providers exist |

#### Phase 4: RC.1 — Feature Parity

| Priority | Title | Description | Assigned | Dependencies |
|----------|-------|-------------|----------|--------------|
| P1 | Versioned configuration support | Implement config versioning (issue #2257) so v1 configs can migrate to v2 CRD format | Batty | Phase 1 |
| P1 | SBOM validation verifier | Port/implement SBOM validation (issue #2353) as a ratify-verifier-go plugin | Batty | ratify-verifier-go |
| P2 | Vulnerability report verifier | Port vuln report verifier (issue #2352) as v2 plugin | Batty | ratify-verifier-go |
| P1 | V2 conformance test suite | Create comprehensive conformance tests covering: multi-verifier, policy enforcement, store fallback, auth flows, CRD reconciliation | Gaff | Phases 1-3 |
| P2 | V1→V2 migration guide | Document upgrade path from v1 to v2 including config migration, CRD changes, plugin migration | Gaff | RC.1 features |

#### Phase 5: GA — Publish

| Priority | Title | Description | Assigned | Dependencies |
|----------|-------|-------------|----------|--------------|
| P0 | Tag ratify-go v1.0.0 | Stabilize API, run full test suite, tag release | Pris | All verifier/store tests pass |
| P0 | Tag ratify v2.0.0 | Final release of K8s provider | Pris | RC.1 complete, conformance pass |
| P1 | Tag ratify-verifier-go notation/cosign v1.0.0 | Release stable verifier plugins | Pris | Integration tests pass |
| P1 | Publish Helm chart for v2 | Update Helm chart repo with v2.0.0 chart | Pris | ratify v2.0.0 tagged |
| P2 | ratify-cli v0.1.0 | Initial CLI release (may defer post-GA) | Pris | Beta.3 if prioritized |
| P1 | Security self-assessment (TSSA) | Complete Tag Security Self Assessment (issue #2035) | Rachael | GA candidate ready |

#### Ongoing / Cross-cutting

| Priority | Title | Description | Assigned | Dependencies |
|----------|-------|-------------|----------|--------------|
| P1 | V1 CVE patching | Continue patching CVEs on v1-dev branch while v2 stabilizes | Leon | None (independent) |
| P2 | V1 maintenance releases | Cherry-pick critical fixes to release-1.4 | Pris | Leon's patches |

### Decision

1. **Immediate focus** (this sprint): Phase 1 — get v2 deployable on K8s. Batty owns all 5 items.
2. **Next sprint**: Phase 2 + Phase 3 Azure auth in parallel. Gaff starts test work; Rachael starts Azure auth.
3. **RC target**: After auth providers + feature parity items land.
4. **GA gate**: Conformance suite passes, TSSA complete, no P0 bugs open.
5. **ratify-cli deferred**: Beta.3 (CLI) is deprioritized — K8s GA is the publish target.

### Rationale

PR #2524's findings show v2 is close to working in K8s but has 5 concrete blockers. Fixing those unblocks all downstream work. Auth is the highest-risk feature gap since most production users need cloud identity integration.

---

## Decision: Alpha.2 E2E Test Plan Structure

**Author:** Gaff
**Date:** 2026-06-01
**Status:** Proposed

### Context
Alpha.2 has 5 deployment blockers being implemented by Batty and Rachael. Tests needed before merge.

### Decision
- Write tests as BATS files following existing `test/bats/` patterns
- Add new CI workflow `e2e-alpha2.yml` as a callable workflow
- Azure auth tests require mock IMDS + WI webhook in kind (not real Azure)
- StoreMux tests need a second local registry instance
- Cert rotation tests rely on Kubernetes Secret update propagation timing (~60s)

### Consequences
- Test infra (mock IMDS, second registry) needs to be built before tests can run
- Cert rotation test timing is non-deterministic — may need generous wait + retry
- Azure auth tests cannot validate real Azure token exchange in CI; mock-only coverage

### Open Questions for Batty/Rachael
1. StoreMux CR field names for prefix-based routing?
2. Health endpoint port — same as external data provider or separate?
3. Cert watcher mechanism — fsnotify or polling?
4. Helm CRDs — `crds/` directory or template-based?

---

## Decision: E2E Test Scaffolds Use Go testing (not BATS)

**Date:** 2026-06-01  
**Author:** Gaff (Tester)  
**Status:** Proposed

### Context

Alpha.2 E2E test scaffolds needed for the 3 remaining deployment blockers (health endpoints, StoreMux fallback, cert rotation). Existing upstream tests use BATS (`test/bats/`).

### Decision

Wrote new E2E tests as Go test files under `src/test/e2e/` using the standard `testing` package and `client-go`. Reasons:

1. Stronger typing — compile-time safety for k8s API interactions
2. Reusable helpers (port-forward, cert generation, log fetching) as Go functions
3. Better IDE support and refactoring
4. Cert rotation tests require crypto operations — natural in Go, awkward in shell

### Trade-offs

- Diverges from existing BATS pattern (adds cognitive overhead for contributors familiar with BATS)
- Requires Go toolchain in CI (already present)
- Could coexist alongside BATS tests long-term

### Impact

Team should align on whether future Alpha.2 tests continue in Go or revert to BATS. This doesn't block implementation — scaffolds can be ported to BATS if team decides against Go E2E tests.

---

## Decision: Azure Auth Provider Design Validated for Alpha.2

**Author:** Rachael  
**Date:** 2026-06-01  
**Status:** Proposed  
**Impacts:** Deckard (arch), Roy (infra/deploy), xinhl (lead)

### Context

The Azure Auth Provider is an Alpha.2 blocker. After reviewing the v2 implementation on `main`, the architecture is sound and structurally fixes the v1 #2504 nil-pointer panic.

### Decision

1. The existing v2 Azure credential provider (`internal/store/credentialprovider/azure/`) is architecturally complete for Alpha.2
2. The critical missing piece is ensuring the `init()` import is present in the v2 binary entrypoint (PR #2524 suggests it may not be)
3. Singleflight for thundering herd protection deferred to Beta
4. No weakening of security guarantees — fail-closed on auth errors

### Action Items

- [ ] Verify `_ "github.com/notaryproject/ratify/v2/internal/store/credentialprovider/azure"` import exists in v2 entrypoint
- [ ] Add integration test with ACR + managed identity in CI
- [ ] Design doc at `azure-auth-design.md` ready for team review

### Risks

- If the import is missing, Azure auth silently doesn't register — config will fail at startup with "credential provider factory of type azure is not registered"

---

## Decision: Azure Auth Provider Implementation Validated

**Date:** 2026-06-01  
**Author:** Rachael (Security Engineer)  
**Status:** Confirmed

### Context

Task was to implement Azure auth provider wiring for v2 and verify the `init()` import in the binary entrypoint (concern from PR #2524).

### Decision

After inspecting upstream `main`, the Azure credential provider (`internal/store/credentialprovider/azure/register.go`) is already fully implemented and the blank import IS present in `cmd/ratify-gatekeeper-provider/register.go`. The PR #2524 concern about missing Azure auth is resolved on main.

### Implementation Notes

- Staged reference implementation in `src/` for review/comparison
- Added explicit nil-credential guard in `exchangeAADTokenForACRToken` (defense-in-depth beyond the structural #2504 fix)
- Added empty-serverAddress guard in `GetWithTTL` (belt-and-suspenders for the nil-pointer pattern)
- Test suite covers: registration, nil-safety, JWT TTL parsing, credential chain configuration

### Impact

No code change needed on main — the wiring is correct. The staged files serve as a validated reference and test harness for the security-critical auth path.

---

*End of decisions log.*
