# Repository Guidelines

## Project Structure & Module Organization

This repository is a small Go 1.22 web service for a navigation/bookmark site. The entry point is `main.go`, which wires configuration, SQLite storage, service logic, and HTTP routes. Domain models live in `internal/domain`, business rules in `internal/service`, persistence in `internal/storage`, runtime flags in `internal/config`, and HTTP handlers/helpers in `internal/transport/http`. Static UI is served from `index.html`. Runtime data belongs in `data/`; `data/sites.json` is the legacy import source and `data/sites.db` is created by the SQLite store. Build output is written to `bin/navigation`.

## Build, Test, and Development Commands

- `go run . -port 8080 -data data`: run the server locally and serve `index.html` plus `/api/*` endpoints.
- `./build.sh`: compile a trimmed binary to `bin/navigation`; override with `OUTPUT=/tmp/navigation ./build.sh`.
- `go test ./...`: run all Go tests. Add tests under the package they cover.
- `go fmt ./...`: format Go source before committing.
- `docker build -t navigation .`: build the container image. Run with a mounted data volume, for example `docker run -p 8080:8080 -v "$PWD/data:/app/data" navigation`.

## Coding Style & Naming Conventions

Use idiomatic Go formatting: tabs via `gofmt`, short package names, exported names only for cross-package APIs, and clear error wrapping or typed errors where callers branch on behavior. Keep layers separated: handlers decode/encode HTTP, services validate and coordinate, storage owns SQL details. Prefer Chinese user-facing messages when matching existing responses and logs. File names should be lowercase with underscores when helpful, such as `sqlite_site_store.go`.

## Testing Guidelines

There is no dedicated test suite yet, so new behavior should include focused `_test.go` files. Prefer table-driven tests for validation and service behavior. For storage tests, use a temporary data directory or SQLite test database so local `data/sites.db` is not modified. Run `go test ./...` before opening a PR.

## Commit & Pull Request Guidelines

Recent history uses short, direct commit messages, including Chinese summaries such as `新增 Dockerfile`. Keep commits concise and scoped to one change. Pull requests should include a brief description, commands run, any API or data migration notes, and screenshots when `index.html` changes. Link related issues when available.

## Security & Configuration Tips

Do not commit generated databases, secrets, or local build artifacts. Keep writable runtime state under `data/`, and validate all API input in the service layer before it reaches storage.
