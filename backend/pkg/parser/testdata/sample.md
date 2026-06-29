# Ragpack

Ragpack is a self-hostable semantic search and retrieval-augmented generation stack
designed to be up and running in minutes. It ships as a single Docker Compose file
with a Go backend, a Next.js admin dashboard, and support for pluggable embedding
and language model providers.

The core idea is simple: you bring your documents and your AI provider, and Ragpack
handles ingestion, chunking, embedding, and retrieval. No vendor lock-in, no
managed service fees, no data leaving your infrastructure.

## Architecture

Ragpack is built around three main components that communicate over an internal
Docker network.

The backend is a Go HTTP server responsible for document ingestion, chunk storage,
vector search, and RAG completions. It exposes a public API on port 9000 and an
admin API on port 9001. The admin port is never published outside the Docker network.

The web admin is a Next.js application that proxies requests to the backend admin
API. It runs on port 3000 and is intended to be accessed via an SSH tunnel when
deployed to a remote server.

The embedder is a pluggable component. Ragpack supports Hugging Face Text
Embeddings Inference for high-throughput production workloads and Ollama for local
development without GPU requirements.

## Data Model

### Collections

A collection is a named group of documents that share an embedding model and vector
dimension. All chunks within a collection live in the same pgvector table. Switching
the embedding model requires creating a new collection and re-ingesting documents,
because vectors from different models are not comparable.

Collections are identified by a URL-safe slug derived from the name at creation time.
The slug is immutable. Renaming a collection changes the display name but not the slug.

### Documents

A document represents a single ingested file. It tracks the source URI, MIME type,
ingestion status, and chunk count. A document can be in one of three states:
ingesting, complete, or failed. Failed documents retain their error message so you
can diagnose and retry.

Documents are deduplicated by source URI within a collection. Re-ingesting the same
URI with the refresh flag will delete the existing chunks and reprocess the file
from scratch.

### Chunks

A chunk is the atomic unit of retrieval. The parser breaks a document into units
based on its structure, and the chunker splits those units into overlapping windows
of a configured token size. Each chunk stores its text, a breadcrumb header for
context, a content hash for deduplication, and the precomputed embedding vector.

At query time Ragpack computes the embedding of the query string and performs a
cosine similarity search against the chunk vectors using pgvector. The top-k results
are returned with their similarity scores.

## Ingestion Pipeline

### Parsing

The parser converts raw file bytes into a sequence of structural units. Each parser
is MIME-type specific. Plain text is split on double newlines. Markdown is split on
heading boundaries with the heading breadcrumb attached as metadata. HTML is
converted to Markdown and then parsed the same way. DOCX, PPTX, and XLSX files are
parsed by walking their internal XML structure.

The parser emits units via a Go iterator so memory usage stays flat regardless of
document size. A ten-thousand-paragraph document consumes the same peak memory as
a ten-paragraph document during parsing.

### Chunking

The chunker receives units from the parser and splits them into fixed-size windows
with configurable overlap. Overlap ensures that context is not lost at chunk
boundaries. The default window is 512 tokens with a 64-token overlap.

Short units that fit within the window are emitted as-is. Long units are split
greedily and each fragment inherits the metadata of the parent unit. This means
every chunk carries the heading breadcrumb of the section it came from, which
improves retrieval relevance for structured documents.

### Embedding

Each chunk is sent to the configured embedding provider and the resulting vector
is stored alongside the chunk text in pgvector. Embedding happens in batches to
reduce round-trip overhead. The batch size is tunable but defaults to 32.

If the embedder is unavailable during ingestion, the job fails with an error and
no partial chunks are committed. Ingestion is all-or-nothing per document.

## Retrieval and RAG

Retrieval takes a query string, embeds it using the same model as the collection,
and runs an approximate nearest-neighbour search with pgvector. The similarity
threshold and result count are configurable per request.

RAG completions pass the retrieved chunks to a language model along with a system
prompt. The prompt template is configurable and supports multiple named presets.
The response includes the generated answer, the chunks that were used, and the
rendered prompt for debugging.

## Deployment

### Requirements

Ragpack requires Docker and Docker Compose. A PostgreSQL instance with the pgvector
extension enabled is required. The provided Compose file includes a pre-configured
Postgres container so no external database is needed for a standard deployment.

For embedding you need either a Hugging Face TEI instance or an Ollama server. TEI
runs natively on Linux x86_64 and under Rosetta emulation on Apple Silicon, though
the emulation path is slow due to ONNX warmup. Ollama is recommended for local
development on Mac.

### Configuration

All configuration is provided through environment variables in a .env file.
The minimum required variables are the database URL, the embedding provider URL,
and the embedding model name. Optional variables control chunk size, overlap,
the default RAG prompt, and admin authentication.

### Upgrading

The CLI reads its own version and passes it as the image tag when starting the
stack. Running ragpack update pulls the images for the current CLI version and
restarts the services. Rolling back is a matter of installing an older CLI version
and running update again.
