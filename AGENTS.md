# Repository Guidelines

## Project Structure & Module Organization

This repository is a Go 1.22 web service plus a Vue 3/Vite frontend for a personal navigation/bookmark and notes site. The entry point is `main.go`, which wires configuration, SQLite storage, note file storage, service logic, and HTTP routes. Domain models live in `internal/domain`, business rules in `internal/service`, persistence in `internal/storage`, runtime flags in `internal/config`, and HTTP handlers/helpers in `internal/transport/http`. Frontend source lives in `frontend/`; production assets are built to `web/dist/` and embedded by Go. Runtime data belongs in `data/`; `data/sites.json` is the legacy import source, `data/sites.db` is created by the SQLite store, and Markdown notes are stored under `data/notes/`. Build output is written to `bin/navigation`.

## Build, Test, and Development Commands

- `go run . -port 8080 -data data`: run the Go API and serve the embedded production frontend from `web/dist/`.
- `cd frontend && npm run dev`: run the Vite dev server; `/api` is proxied to `http://localhost:8080`.
- `cd frontend && npm run build`: type-check and build frontend assets into `web/dist/`.
- `./build.sh`: run the frontend build, then compile a trimmed binary to `bin/navigation`; override with `OUTPUT=/tmp/navigation ./build.sh`.
- `go test ./...`: run all Go tests. Add tests under the package they cover.
- `go fmt ./...`: format Go source before committing.
- `cd frontend && npm run lint`: run Vue/TypeScript type checking without emitting files.
- `docker build -t navigation .`: build the container image. Run with a mounted data volume, for example `docker run -p 8080:8080 -v "$PWD/data:/app/data" navigation`.

## Coding Style & Naming Conventions

Use idiomatic Go formatting: tabs via `gofmt`, short package names, exported names only for cross-package APIs, and clear error wrapping or typed errors where callers branch on behavior. Keep layers separated: handlers decode/encode HTTP, services validate and coordinate, storage owns SQL and file-system details. Prefer Chinese user-facing messages when matching existing responses and logs. File names should be lowercase with underscores when helpful, such as `sqlite_site_store.go` and `note_file_store.go`. For frontend code, follow the existing Vue single-file component pattern, composables under `frontend/src/composables`, shared API/types under `frontend/src/lib` and `frontend/src/types`, and Tailwind utility styling in the current design language.

## Testing Guidelines

The repository has focused Go tests for service, handler, SQLite note storage, and note file storage behavior. New backend behavior should include focused `_test.go` files in the package it covers. Prefer table-driven tests for validation and service behavior. For storage tests, use a temporary data directory or SQLite test database so local `data/sites.db` and `data/notes/` are not modified. Run `go test ./...` before opening a PR. For frontend changes, run `cd frontend && npm run lint`, and run `cd frontend && npm run build` when changing production UI behavior or embedded assets.

## Commit & Pull Request Guidelines

Recent history uses short, direct commit messages, including Chinese summaries such as `新增 Dockerfile`. Keep commits concise and scoped to one change. Pull requests should include a brief description, commands run, any API or data migration notes, and screenshots when `frontend/` or `web/dist/` changes. Link related issues when available.

## Security & Configuration Tips

Do not commit generated databases, SQLite WAL/SHM files, note content, secrets, dependency folders, caches, or local build artifacts. Keep writable runtime state under `data/`, and validate all API input in the service layer before it reaches storage. `go-sqlite3` requires CGO and a local C toolchain for builds and tests.
