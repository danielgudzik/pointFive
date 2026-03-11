run:
	go run .

test:
	go test ./... -v

# Quick smoke test against a running server (requires: make run in another terminal)
demo:
	@bash -c '\
	  echo "--- submit job ---"; \
	  RESP=$$(curl -s -X POST localhost:8080/jobs \
	    -H "Content-Type: application/json" \
	    -d '"'"'{"items":[{"id":"1","payload":{"name":"alice","score":10}},{"id":"2","payload":{"name":"bob","score":5}}]}'"'"'); \
	  echo "$$RESP" | jq .; \
	  JOB_ID=$$(echo "$$RESP" | jq -r ".id"); \
	  echo "\n--- wait for processing ---"; \
	  sleep 1; \
	  echo "\n--- get results ---"; \
	  curl -s localhost:8080/jobs/$$JOB_ID | jq .; \
	  echo "\n--- health ---"; \
	  curl -s localhost:8080/health | jq .; \
	'

.PHONY: run test demo
