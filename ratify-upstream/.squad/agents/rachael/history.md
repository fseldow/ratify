# Rachael — History

## Session Log
- Team created. Awaiting first task.
- 2026-06-01: Completed Azure Auth Provider design doc for Alpha.2. Analyzed v2 credential provider architecture on `main` branch, validated it addresses v1 #2504 panic structurally. Design written to `azure-auth-design.md`.

## Learnings
- 2026-06-01: Confirmed Azure import IS present in `cmd/ratify-gatekeeper-provider/register.go` on main — PR #2524 concern was unfounded for the gatekeeper provider binary. The import `_ "github.com/notaryproject/ratify/v2/internal/store/credentialprovider/azure"` is already wired.
- 2026-06-01: Created staging implementation in `src/` matching upstream patterns. Key structural safety: `serverAddress` is a direct parameter to `GetWithTTL()`, not a stored getter field — this is the #2504 nil-pointer fix by design.
- 2026-06-01: `exchangeAADTokenForACRToken` extracted as package-level function with explicit nil-credential guard for testability.
- v2 Azure auth already exists on `main` in `internal/store/credentialprovider/azure/` — it's not greenfield, it's validation + wiring work
- v2 uses a factory pattern (`credmanager.go`) with `init()` registration — Azure auth is opt-in via import side-effects in the entrypoint binary
- `CachedProvider` wraps all credential sources with in-memory TTL cache (capacity 10, keyed by serverAddress)
- Credential chain order: WorkloadIdentity → ManagedIdentity (user-assigned if clientID set, else system-assigned)
- The #2504 v1 bug (nil pointer panic) is structurally prevented in v2 because `serverAddress` is passed directly to `GetWithTTL()` rather than stored in a lazily-initialized getter
- StoreMux routing is scope-based — credentials are per-store, wired at construction time, not dynamically selected
- PR #2524 explicitly notes "no Azure auth providers" in v2 — this means the `init()` import may be missing from the v2 binary entrypoint; needs verification
- No singleflight protection on cache miss — thundering herd possible under load; acceptable for Alpha.2, fix for Beta

## Sprint Completion (2026-06-01)
- ✅ Azure auth wiring confirmed and validated:
  - Blank import `_ "github.com/notaryproject/ratify/v2/internal/store/credentialprovider/azure"` IS present in v2 entrypoint
  - PR #2524 concern resolved — no code change needed on main
  - Added nil-credential guard + empty-serverAddress guard to staged reference implementation
- ✅ Test suite created covering: registration, nil-safety, JWT TTL parsing, credential chain
- ✅ Design doc: `azure-auth-design.md` ready for team review
- 🔄 Awaiting integration test implementation (ACR + managed identity in CI)
- Cross-team: Batty implemented 3 blockers; Gaff deferred Azure tests to Go-based E2E scaffolds
