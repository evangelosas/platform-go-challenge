# GlobalWebIndex Engineering Challenge

## Running

- Go 1.20+
- Start the server (in-memory store):
  - `go run .`
- Start with file persistence:
  - `FAV_STORE=file FAV_PATH=./data/favourites.json go run .`
  - By default, file store path is `data/favourites.json` if FAV_PATH is not set.

## API

Base path: `/users/{userID}/favourites`

- GET `/users/{userID}/favourites`
  - Query params:
    - `limit` (int, default 100, max 1000)
    - `offset` (int, default 0)
    - `sort` one of `created_at`, `type` (default `created_at`)
    - `order` `asc` or `desc` (default `asc`)
- POST `/users/{userID}/favourites`
  - Body: Add a single asset
  - Example:
    - `{"type":"chart","description":"My chart","payload":{"title":"Sales","xAxisTitle":"Month","yAxisTitle":"USD","data":[1,2,3]}}`
- POST `/users/{userID}/favourites/bulk`
  - Body: JSON array of the same objects as above to add in bulk
- PATCH `/users/{userID}/favourites/{assetID}`
  - Body: `{"description":"new description"}`
- DELETE `/users/{userID}/favourites/{assetID}`

Asset response objects contain common fields `id`, `type`, `description`, `createdAt` plus type-specific fields.

## Storage

Two store implementations:
- In-memory (default): fast, volatile.
- File-backed (JSON): persists to a single file. Configure via env vars:
  - `FAV_STORE=file`
  - `FAV_PATH=./data/favourites.json`

## Tests and Benchmarks

- Run tests: `go test ./...`
- Run benchmarks: `go test -bench=. -benchmem`

## Submission

Just create a fork from the current repo and send it to us!

Good luck, potential colleague!
