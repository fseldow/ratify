# Gaff — History

## Session Log
- Team created. Awaiting first task.
- 2026-06-01: Wrote Alpha.2 E2E test plan covering all 5 deployment blockers (health probes, Helm CRDs, StoreMux, Azure auth, cert rotation). Delivered to `alpha2-test-plan.md`.
- 2026-06-01: Updated test plan — removed CRD and Azure Auth scenarios (deferred to Beta). Renumbered to 3 sections, 9 scenarios. Wrote Go E2E test scaffolds under `src/test/e2e/` for health endpoints, StoreMux fallback, and cert rotation.

## Learnings
- Ratify E2E tests use BATS framework (`test/bats/`) with `helpers.bash` providing `assert_success`/`assert_failure`.
- CI uses kind clusters with Gatekeeper, orchestrated via `make e2e-bootstrap`, `make e2e-deploy-ratify`, `make test-e2e`.
- Existing Azure tests live in `azure-test.bats` — new Azure auth tests should extend or parallel that file.
- The `e2e-k8s.yml` workflow is a reusable callable workflow pattern (uses `workflow_call`). Alpha.2 tests should follow same pattern.
- Registry at `registry:5000` is available in CI for local image testing. StoreMux tests will need a second registry instance.
- CRD testing requires careful attention to Helm 3's `crds/` directory semantics vs template-based CRDs — behavior differs on uninstall.
- 2026-06-01: Helm CRD packaging tests deferred to Beta per user directive. Test plan reduced from 15 to 9 scenarios (3 blockers × 3 tests each).
- 2026-06-01: E2E test scaffolds written as Go test files under `src/test/e2e/` using standard `testing` package + client-go, rather than BATS. This allows stronger typing, better IDE support, and reuse of Go helper functions.
- 2026-06-01: Cert rotation tests rely on kubelet Secret sync period (~60s) — tests need generous timeouts (120s+) to avoid flakiness.
- 2026-06-01: StoreMux fallback behavior uses `RegisterFallback` API in ratify-go library — single-store-no-scope config should trigger this path automatically.

## Sprint Completion (2026-06-01)
- ✅ Test plan finalized:
  - Removed CRD scenarios (deferred to Beta per user directive)
  - Removed Azure Auth scenarios (Rachael validated on main)
  - Reduced to 3 blockers × 3 scenarios = 9 total tests
- ✅ E2E test scaffolds created under `src/test/e2e/` (Go-based, not BATS):
  - Health endpoint tests (flat paths, readiness signals)
  - StoreMux fallback tests (wildcard scope `"*"`)
  - Cert rotation tests (fsnotify + polling resilience)
- 🔄 Decision filed: Go vs BATS approach for future Alpha.2 tests (team alignment needed)
- 🔄 Test infrastructure requirements: mock IMDS, second registry, cert rotation timing (120s+ waits)
- Cross-team: Batty implemented blockers; Rachael validated Azure auth; decisions documented in `decisions.md`
