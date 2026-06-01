# Batty — History

## Session Log
- Team created. Awaiting first task.
- 2026-06-01: Produced alpha2-design.md covering 4 deployment blockers for v2.0.0-alpha.2 (health endpoints, Helm CRD packaging, StoreMux fallback routing, cert rotation watcher). Reviewed PR #2524, upstream codebase structure, ratify-go StoreMux, and existing tlssecret watcher.

## Learnings
- 2026-06-01: Implemented all 3 alpha.2 blockers as staged code under `src/`. Health endpoints use atomic bools for lock-free probe responses. StoreMux fallback uses `RegisterFallback()` from ratify-go. Cert watcher combines fsnotify + SHA-256 polling fallback for k8s symlink rotation resilience.
- The upstream `server` struct uses an embedded `ServerOptions` — health field goes alongside `sfGroup` as a pointer to keep the zero-value safe.
- `ratify-go` StoreMux.RegisterFallback() is the canonical way to handle catch-all stores; the `"*"` scope is a user-facing config convention that maps to it.
- For cert rotation in k8s, fsnotify alone is unreliable with atomic symlink swaps (`..data` → timestamped dir). Polling SHA-256 every 30s is the safety net.
- v2 server uses gorilla/mux at `internal/httpserver/server.go`; straightforward to add new GET routes for health checks.
- `ratify-go` `StoreMux` already has `RegisterFallback()` but it's never called from the ratify v2 `internal/store/factory.go` — that's the root cause of the "no matching store" issue with wildcard patterns.
- The existing `internal/httpserver/tlssecret/certwatcher.go` is a solid pattern (fsnotify + atomic.Pointer) to reuse for trust store cert watching.
- Helm CRDs in `crds/` dir are never upgraded on `helm upgrade` — must use hook-based Job pattern (cert-manager style) for alpha iteration.
- The v2 Helm chart lives at `deployments/ratify-gatekeeper-provider/` not `charts/ratify/` (v1 path no longer exists on main).
- ScopedExecutor routing (executor.go) mirrors StoreMux: wildcard→registry→repository precedence. Both need fallback support for `"*"` scope.

## Sprint Completion (2026-06-01)
- ✅ All 3 Alpha.2 blockers implemented and documented:
  - Health endpoints: `/healthz`, `/readyz` at flat paths, no auth wrapping
  - StoreMux fallback: `"*"` scope convention for global fallback via `RegisterFallback()`
  - Cert watcher: dual fsnotify + SHA-256 polling (30s) with 2s debounce
- ✅ 5 files created under `src/` (health, store factory, executor, verifier watcher)
- 🔄 Awaiting team review of `alpha2-design.md`
- 🔄 Awaiting Deckard approval for main merge
- Cross-team: Rachael validated Azure auth wiring; Gaff created test scaffolds (Go-based)
