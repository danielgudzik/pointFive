# API Reference

Base URL: `http://localhost:8080`

All requests and responses use JSON. Error responses always include an `"error"` field.

---

## GET /health

Health check endpoint.

**Response `200 OK`**

```json
{"status": "ok"}
```

---

## POST /jobs

Submit a batch job for asynchronous processing.

**Request body**

| Field   | Type    | Required | Description                       |
|---------|---------|----------|-----------------------------------|
| `items` | array   | yes      | List of items to process          |

Each item:

| Field     | Type   | Required | Description                          |
|-----------|--------|----------|--------------------------------------|
| `id`      | string | yes      | Unique identifier for the item       |
| `payload` | object | yes      | Arbitrary key-value data to process  |

**Example request**

```bash
curl -s -X POST http://localhost:8080/jobs \
  -H 'Content-Type: application/json' \
  -d '{
    "items": [
      {"id": "a", "payload": {"name": "alice", "score": 10}},
      {"id": "b", "payload": {"name": "bob",   "score": 5}}
    ]
  }'
```

**Response `202 Accepted`**

```json
{
  "id":         "20060102150405.999999999",
  "status":     "pending",
  "items":      [...],
  "created_at": "2024-01-01T00:00:00Z"
}
```

The job is queued for processing immediately. Use the returned `id` to poll for results.

---

## GET /jobs/{id}

Retrieve the status and results of a previously submitted job.

**Path parameter**

| Parameter | Description        |
|-----------|--------------------|
| `id`      | Job ID from POST   |

**Example request**

```bash
curl -s http://localhost:8080/jobs/20060102150405.999999999
```

**Response `200 OK` — job pending**

```json
{
  "id":         "20060102150405.999999999",
  "status":     "pending",
  "items":      [...],
  "created_at": "2024-01-01T00:00:00Z"
}
```

**Response `200 OK` — job done**

```json
{
  "id":     "20060102150405.999999999",
  "status": "done",
  "items":  [...],
  "results": [
    {"item_id": "a", "output": {"name": "[processed] alice", "score": 20}},
    {"item_id": "b", "output": {"name": "[processed] bob",   "score": 10}}
  ],
  "created_at": "2024-01-01T00:00:00Z",
  "done_at":    "2024-01-01T00:00:00.5Z"
}
```

**Response `404 Not Found`**

```json
{"error": "job not found"}
```

---

## Processing logic

The pipeline transforms each item's payload field-by-field:

| Payload value type | Transformation          |
|--------------------|-------------------------|
| `string`           | Prefixed with `"[processed] "` |
| `number`           | Multiplied by 2         |
| Other              | Passed through unchanged|
