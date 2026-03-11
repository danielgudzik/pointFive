Add a new HTTP endpoint to this project: $ARGUMENTS

Steps:
1. Read `api/handlers.go` and `api/server.go`
2. Add a handler method on the `handlers` struct in `api/handlers.go`:
   - Decode JSON body: `json.NewDecoder(r.Body).Decode(&body)` with inline struct
   - Path params: `r.PathValue("name")`
   - Success: `writeJSON(w, http.StatusXxx, value)`
   - Error: `writeError(w, http.StatusXxx, "message")`
3. Register in `NewServer()` in `api/server.go`: `mux.HandleFunc("METHOD /path", h.method)`
4. Run `make test` to verify nothing broke