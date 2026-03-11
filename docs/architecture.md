# Architecture

## Overview

pointFive is structured in two layers:

```
HTTP API (api/)
    └── Pipeline (pipeline/)
```

`main.go` wires the two together and owns the process lifecycle.

---

## Request flow

```
Client
  │
  ▼
api/server.go       — route registration, logging middleware
  │
  ▼
api/handlers.go     — decode request, call pipeline, encode response
  │
  ▼
pipeline/pipeline.go — store job, fan-out to workers, collect results
```

1. A `POST /jobs` request reaches `SubmitJob` in `handlers.go`.
2. The handler builds a `pipeline.Job` and calls `pipeline.Submit`.
3. `Submit` stores the job (status `"pending"`) and launches a goroutine to process it asynchronously.
4. The handler immediately returns `202 Accepted` with the job ID.
5. The client polls `GET /jobs/{id}` until status is `"done"`.

---

## Worker pool (fan-out / fan-in)

```
Submit(job)
  │
  └─► goroutine: process(job)
            │
            ├─ items channel ◄──────────────── all items enqueued
            │
            ├─ worker 1 ──► processItem ──► results channel
            ├─ worker 2 ──► processItem ──► results channel
            ├─ worker N ──► processItem ──► results channel
            │
            └─ collect results, mark job "done"
```

- **Fan-out:** N workers (`WorkerCount`, default 4) each read from the shared items channel.
- **Fan-in:** Each worker writes its result to a shared results channel.
- A `sync.WaitGroup` ensures the result collector waits until every worker has finished before the job is marked done.
- Both channels are buffered to the size of the item list, so enqueueing never blocks.

---

## Job store

Jobs are held in a `map[string]*Job` inside the `Pipeline` struct, protected by a `sync.RWMutex`.

- Reads (`Get`) acquire a read lock.
- Writes (`Submit`, result collection, status update) acquire a write lock.

There is no persistence; jobs are lost on restart.

---

## Graceful shutdown

`main.go` listens for `SIGINT`/`SIGTERM`. On signal:

1. A derived context is cancelled (10-second timeout).
2. `http.Server.Shutdown` drains in-flight requests.
3. In-progress pipeline jobs receive context cancellation signals; each `processItem` call receives the context and can respect cancellation.

---

## Key types

```
pipeline.Item    — {ID string, Payload map[string]any}
pipeline.Result  — {ItemID string, Output map[string]any, Error string}
pipeline.Job     — {ID, Status, Items, Results, CreatedAt, DoneAt}
pipeline.Pipeline— owns the worker pool logic and job store
api.Server       — owns the http.Server and route registration
```
