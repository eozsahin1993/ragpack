---
sidebar_position: 1
---

# Getting Started

RagPack is self-hostable semantic search and RAG infrastructure. This guide gets you from zero to a running instance in a few minutes.

## Prerequisites

- [Docker](https://docs.docker.com/get-docker/) and Docker Compose
- Node.js 18+ (for the CLI)

## Install the CLI

```bash
npm install -g ragpack
```

Or use it without installing:

```bash
npx ragpack <command>
```

## Quick start

**1. Initialize your config:**

```bash
npx ragpack init
```

This creates `.env.ragpack` in the current directory.

**2. Start the stack:**

```bash
# With Ollama (fully local)
npx ragpack start --profile ollama

# With OpenAI — edit .env.ragpack first
npx ragpack start
```

**3. Open the admin UI:**

Go to [http://localhost:3000](http://localhost:3000) to create collections, ingest documents, and run queries.

The REST API is available at `http://localhost:9000/api/v1`.

## Configuration

Edit `.env.ragpack` to configure your embedding provider:

```env
# Embedding provider: ollama | openai | tei
EMBED_PROVIDER=ollama

# Ollama
OLLAMA_BASE_URL=http://ollama:11434
OLLAMA_EMBED_MODEL=nomic-embed-text

# OpenAI
OPENAI_API_KEY=sk-...
OPENAI_EMBED_MODEL=text-embedding-3-small

# Ingestion
WORKER_COUNT=5
EMBED_RATE_LIMIT=10
```

## CLI reference

| Command | Description |
|---|---|
| `ragpack init` | Create `.env.ragpack` in the current directory |
| `ragpack start [--profile ollama]` | Start the stack |
| `ragpack stop [-v]` | Stop the stack (`-v` removes volumes and all data) |
| `ragpack logs [service]` | Tail logs (`backend`, `web-admin`, `ollama`) |
| `ragpack update` | Pull latest images and restart |
| `ragpack eject` | Copy `docker-compose.yml` locally for customization |
