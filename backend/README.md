# RagPack Backend

Go + Fiber HTTP server for the RagPack engine. Uses LanceDB (via CGO) for vector storage and SQLite for metadata.

## Requirements

- Go 1.21+
- macOS arm64 (darwin_arm64) — Linux support coming
- curl (for downloading the LanceDB native library on first run)

## Start dev server

```bash
cd backend
./dev.sh
```

That's it. On first run it downloads the LanceDB native library automatically. The server starts on `http://localhost:9000`.

## API

All routes are prefixed with `/api/v1`.

### Health

```
GET /api/v1/health
```

### Collections

```
POST   /api/v1/collections                  Create a collection
GET    /api/v1/collections                  List all collections
GET    /api/v1/collections/:name            Get by name
GET    /api/v1/collections/id/:id           Get by ID
DELETE /api/v1/collections/:name            Delete by name
DELETE /api/v1/collections/id/:id           Delete by ID
```

**Create body:**
```json
{
  "name": "company_wiki",
  "embed_model": "text-embedding-3-small",
  "vector_dim": 1536
}
```

### Jobs

```
GET /api/v1/collections/:name/jobs                    List jobs for a collection
GET /api/v1/collections/:name/jobs/status/:status     List by status (pending | processing | complete | failed)
GET /api/v1/collections/:name/jobs/:id                Get a job
```

### Ingest

```
POST /api/v1/collections/:name/ingest       (coming soon)
```

### Query

```
POST /api/v1/collections/:name/query        (coming soon)
```

## Manual build

If you want to build without `dev.sh`, export the CGO flags first:

```bash
LANCEDB_MODULE="$(go env GOPATH)/pkg/mod/github.com/lancedb/lancedb-go@v0.1.2"

export CGO_CFLAGS="-I${LANCEDB_MODULE}/include"
export CGO_LDFLAGS="${LANCEDB_MODULE}/lib/darwin_arm64/liblancedb_go.a -framework Security -framework CoreFoundation"

go build ./...
```
