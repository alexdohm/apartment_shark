## Formatting & linting
	•	Always run gofmt, goimports, go vet ./... before committing.
	•	Add staticcheck ./... for deeper style and bug checks.
## Project layout
	•	Root-level go.mod.
	•	/cmd/<service>/main.go for entry-points.
	•	/internal/... for private packages.
	•	/pkg/... for shared libraries.
## Modules & builds
	•	Pin to the latest stable Go (e.g., go 1.24) in go.mod.
	•	Run go mod tidy after changing dependencies.
	•	Build with go build -trimpath -ldflags "-s -w" for small, reproducible binaries.
## Error handling
	•	Return errors; wrap with %w.
	•	Use errors.Is/As for checks.
	•	Avoid panics outside main() or tests.
## Context & cancellation
	•	First arg of any I/O or long-running function is ctx context.Context.
	•	Respect <-ctx.Done() or ctx.Err() in loops.
## Concurrency
	•	Protect shared state with channels or sync.Mutex.
	•	Use errgroup.Group for parallel fan-out + aggregated errors.
	•	Run go test -race frequently.
## Testing
	•	Write table-driven sub-tests in _test.go.
	•	Track coverage via go test ./... -cover.
## Logging
	•	Use log/slog for structured logs.
## Security & hardening
	•	Run govulncheck ./... (via go vet -vettool).
	•	Build & test with -race; enable sanitizers where possible.