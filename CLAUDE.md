# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

# pointFive — Interview Context

## Architecture
```
main.go → api.NewServer() → handlers.go        (HTTP layer)
                                 ↓
                         pipeline.ItemPipeline  (thin wiring layer)
                                 ↓
                         utils/workerpool       (generic worker pool, job store)
                                 ↓
                         processItem()          (← data transform logic lives here)
```

## Key files
| File | Role |
|------|------|
| `main.go` | Wires config → pipeline → HTTP server, signal handling |
| `config/config.go` | `AppConfig`, `Load()` — Viper-backed config loader (env vars + `.env`) |
| `config/env.go` | Env var key constants (`SERVER_ADDR`, `WORKER_COUNT`, etc.) |
| `api/server.go` | Mux setup, route registration, `withLogging` middleware |
| `api/handlers.go` | Handlers, `writeJSON`/`writeError` helpers, `newID()` |
| `utils/workerpool/workerpool.go` | Generic `Pipeline[In,Out]` + `Job[In,Out]` — portable worker pool infra |
| `pipeline/item_pipeline.go` | `ItemPipeline` type alias + `NewItemPipeline()` constructor |
| `pipeline/item_processor.go` | `processItem()` — data transform business logic (extend here) |
| `pipeline/item_pipeline_test.go` | Integration test: submit + get |
| `pipeline/item_processor_test.go` | Unit tests for `processItem` transforms |
| `entities/job.go` | Domain types: `ItemJob` (alias), `Item`, `Result` |
| `entities/pipeline_settings.go` | `PipelineSettings` (worker count + logger) |
| `.env` | Default runtime values loaded by Viper on startup |

## Extension points
- **New endpoint**: add method to `handlers` struct → register in `NewServer()` mux
- **New data transform**: edit `processItem()` in `pipeline/item_processor.go`
- **New job type** (e.g. ImageJob): (1) add `ImageInput`, `ImageOutput`, `type ImageJob = wp.Job[ImageInput, ImageOutput]` in `entities/`; (2) add `pipeline/image_pipeline.go` with `ImagePipeline` alias + `NewImagePipeline()`; (3) add `pipeline/image_processor.go` with `processImage()`; (4) add handler + route in `api/`
- **New config value**: add const to `config/env.go`, add field to `AppConfig` in `config/config.go` (SetDefault + GetXxx), add to `.env`
- **Struct placement**: all structs belong in `entities/`; non-struct code (handlers, logic, server wiring) lives in its own package. Exception: `api.Config` stays in `api/` to avoid import cycles (`*pipeline.ItemPipeline` field).

## Conventions (must follow)
- Responses: `writeJSON(w, http.StatusXxx, val)` / `writeError(w, http.StatusXxx, "msg")`
- Logging: `h.log.Info("msg", "key", val)` via `log/slog`
- Thread safety: `p.mu.Lock()` writes, `p.mu.RLock()` reads on `p.jobs`
- Route syntax: `"METHOD /path/{param}"` (Go 1.22 mux); path params via `r.PathValue("name")`
- Config: `config.Load()` returns `*AppConfig`; use `config.SERVER_ADDR` etc. constants for all env var keys
- External dependency: `github.com/spf13/viper` (config loading only)

## Current endpoints
| Method | Path | Handler |
|--------|------|---------|
| GET | `/health` | `health` |
| GET | `/jobs` | `listJobs` |
| POST | `/jobs` | `submitJob` |
| GET | `/jobs/{id}` | `getJob` |

## Commands
- `make run`  — start server on :8080
- `make test` — run all tests verbose
- `make demo` — smoke test (needs server running in another terminal)
- `go test ./pipeline/... -run TestName` — run a single test by name

## When you make changes, update these files

| Change made | Files to update |
|-------------|-----------------|
| Add/remove/rename an endpoint | `CLAUDE.md` key files table + extension points |
| Add a new struct field or type | `CLAUDE.md` extension points section |
| Add/change a convention (logging, errors, etc.) | `CLAUDE.md` conventions section |
| Add a new make target | `CLAUDE.md` commands section |
| Add/remove/rename an env var | `config/env.go` + `config/config.go` + `.env` + `README.md` Configuration table |
| Any of the above | `memory/MEMORY.md` if the pattern is captured there |

After implementing any feature, check: does CLAUDE.md still accurately describe the project?
If not, update it before finishing.
