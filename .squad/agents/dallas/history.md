# Dallas — History

## Project Context
- **Project:** akssec — AKS image integrity
- **Stack:** Go, Kubernetes, admission controllers
- **Focus:** Container image signing, verification, admission policies
- **User:** xinhl

## Team Updates

### 2026-06-02: Kane E2E Framework Changes
- E2E tests now use BATS (reusing v1 approach) instead of Go
- v2 Executor CRD coverage added via `e2e-k8s-v2.yml` workflow
- No production logic changes to deployment/Helm chart
- Chart now includes `stores[0].plainHttp` support for e2e registry at `registry:5000`

## Learnings
