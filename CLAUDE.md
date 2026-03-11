# pointFive — Interview Context

## Architecture
```
main.go → api.NewServer() → handlers.go  (HTTP layer)
                                 ↓
                         pipeline.Pipeline  (worker pool, job store)
                                 ↓
                         processItem()     (← data transform logic lives here)
```

## Key files
| File | Role |
|------|------|
| `main.go` | Wires pipeline + HTTP server, signal handling |
| `api/server.go` | Mux setup, route registration, `withLogging` middleware |
| `api/handlers.go` | Handlers, `writeJSON`/`writeError` helpers, `newID()` |
| `pipeline/pipeline.go` | Worker pool, `map[string]*Job` store (RWMutex), `processItem` |
| `pipeline/pipeline_test.go` | Unit tests — follow these patterns for new tests |

## Extension points
- **New endpoint**: add method to `handlers` struct → register in `NewServer()` mux
- **New data transform**: edit `processItem()` in `pipeline/pipeline.go`
- **New job field**: add to `Job` struct, populate in `Submit()` or `process()`

## Conventions (must follow)
- Responses: `writeJSON(w, http.StatusXxx, val)` / `writeError(w, http.StatusXxx, "msg")`
- Logging: `h.log.Info("msg", "key", val)` via `log/slog`
- Thread safety: `p.mu.Lock()` writes, `p.mu.RLock()` reads on `p.jobs`
- Route syntax: `"METHOD /path/{param}"` (Go 1.22 mux); path params via `r.PathValue("name")`
- No external dependencies — stdlib only

## Commands
- `make run`  — start server on :8080
- `make test` — run all tests verbose
- `make demo` — smoke test (needs server running in another terminal)

## When you make changes, update these files

| Change made | Files to update |
|-------------|-----------------|
| Add/remove/rename an endpoint | `CLAUDE.md` key files table + extension points |
| Add a new struct field or type | `CLAUDE.md` extension points section |
| Add/change a convention (logging, errors, etc.) | `CLAUDE.md` conventions section |
| Add a new make target | `CLAUDE.md` commands section |
| Any of the above | `memory/MEMORY.md` if the pattern is captured there |

After implementing any feature, check: does CLAUDE.md still accurately describe the project?
If not, update it before finishing.
