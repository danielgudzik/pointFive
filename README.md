# pointFive

A concurrent batch job processing service with a REST API, built in Go.

pointFive accepts batch jobs containing multiple data items, processes them concurrently using a configurable worker pool, and exposes job status and results over HTTP.

## Features

- Worker pool-based concurrent processing (fan-out/fan-in pattern)
- In-memory job store with thread-safe access
- REST API for job submission and result retrieval
- Graceful shutdown with context propagation
- Structured logging via `slog`

## Getting Started

**Requirements:** Go 1.22+

```bash
# Run the server (listens on :8080)
make run

# Run tests
make test

# Run a smoke-test demo (requires server running)
make demo
```

## API

See [docs/api.md](docs/api.md) for the full API reference.

| Method | Path         | Description             |
|--------|--------------|-------------------------|
| GET    | `/health`    | Health check            |
| POST   | `/jobs`      | Submit a batch job      |
| GET    | `/jobs/{id}` | Get job status/results  |

### Submit a job

```bash
curl -s -X POST http://localhost:8080/jobs \
  -H 'Content-Type: application/json' \
  -d '{
    "items": [
      {"id": "1", "payload": {"name": "alice", "score": 10}},
      {"id": "2", "payload": {"name": "bob",   "score": 5}}
    ]
  }'
```

Response (`202 Accepted`):

```json
{
  "id": "20060102150405.999999999",
  "status": "pending",
  "items": [...],
  "created_at": "2024-01-01T00:00:00Z"
}
```

### Get results

```bash
curl -s http://localhost:8080/jobs/<id>
```

Response (`200 OK`):

```json
{
  "id": "20060102150405.999999999",
  "status": "done",
  "results": [
    {"item_id": "1", "output": {"name": "[processed] alice", "score": 20}},
    {"item_id": "2", "output": {"name": "[processed] bob",   "score": 10}}
  ],
  "created_at": "2024-01-01T00:00:00Z",
  "done_at":    "2024-01-01T00:00:00.5Z"
}
```

## Configuration

Configuration is set in `main.go`:

| Field          | Default | Description                       |
|----------------|---------|-----------------------------------|
| `WorkerCount`  | `4`     | Number of concurrent workers      |
| `Addr`         | `:8080` | HTTP listen address               |

## Project Structure

```
.
├── main.go          # Entry point, wires pipeline and HTTP server
├── api/
│   ├── server.go    # HTTP server setup and route registration
│   └── handlers.go  # Request handlers
└── pipeline/
    ├── pipeline.go      # Worker pool, job store, data transformation
    └── pipeline_test.go # Unit tests
```

## Architecture

See [docs/architecture.md](docs/architecture.md) for a detailed walkthrough of the pipeline design and concurrency model.