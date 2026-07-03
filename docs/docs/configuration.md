---
sidebar_position: 3
---

# Configuration

All configuration is done via `.env.ragpack`, created by `ragpack init`. The file is read on startup — restart the stack after making changes.

```bash
ragpack stop && ragpack start
```

## Embedding providers

RagPack needs an embedding model to index and query documents. Set `DEFAULT_EMBED_PROVIDER` to one of `ollama`, `openai`, or `tei`.

### Ollama (default, fully local)

```env
DEFAULT_EMBED_PROVIDER=ollama
OLLAMA_BASE_URL=http://ollama:11434
OLLAMA_EMBED_MODEL=nomic-embed-text
```

Start the stack with the Ollama profile to include the Ollama service:

```bash
ragpack start --profile ollama
```

### OpenAI

```env
DEFAULT_EMBED_PROVIDER=openai
OPENAI_API_KEY=sk-...
OPENAI_EMBED_MODEL=text-embedding-3-small
```

### HuggingFace TEI (self-hosted)

```env
DEFAULT_EMBED_PROVIDER=tei
TEI_BASE_URL=http://localhost:8080
TEI_EMBED_MODEL=BAAI/bge-small-en-v1.5
```

## LLM providers

An LLM is required to use the RAG endpoint (`collection.rag()`). Set `DEFAULT_LLM_PROVIDER` to one of `openai`, `ollama`, or `anthropic`.

If no LLM provider is configured, semantic search (`collection.findSimilar()`) still works — only RAG is unavailable.

### OpenAI

```env
DEFAULT_LLM_PROVIDER=openai
OPENAI_API_KEY=sk-...
OPENAI_LLM_MODEL=gpt-4o-mini
```

### Ollama

```env
DEFAULT_LLM_PROVIDER=ollama
OLLAMA_BASE_URL=http://ollama:11434
OLLAMA_LLM_MODEL=llama3.2
```

### Anthropic

```env
DEFAULT_LLM_PROVIDER=anthropic
ANTHROPIC_API_KEY=sk-ant-...
ANTHROPIC_MODEL=claude-haiku-4-5-20251001
```

## Ingestion

```env
WORKER_COUNT=5        # concurrent ingestion workers
EMBED_RATE_LIMIT=10   # max embedding API calls per second
MAX_UPLOAD_SIZE_MB=20 # max file upload size
```

`EMBED_RATE_LIMIT` is especially important when using an external provider like OpenAI — set it to stay within your API tier's rate limit.

## Chunking defaults

These apply to all collections unless overridden at collection creation time.

```env
CHUNK_STRATEGY=auto   # auto | paragraph | sliding_window | section | unit | row_group
CHUNK_SIZE=2000       # max characters per chunk
CHUNK_OVERLAP=200     # characters carried into the next chunk
```

## RAG defaults

```env
DEFAULT_PROMPT_SLUG=basic_rag  # prompt template used when none is specified
```

## Full reference

| Variable | Default | Description |
|---|---|---|
| `DEFAULT_EMBED_PROVIDER` | `ollama` | Embedding provider: `ollama` \| `openai` \| `tei` |
| `OLLAMA_BASE_URL` | `http://ollama:11434` | Ollama server URL |
| `OLLAMA_EMBED_MODEL` | `nomic-embed-text` | Ollama embedding model |
| `OPENAI_API_KEY` | — | OpenAI API key |
| `OPENAI_EMBED_MODEL` | `text-embedding-3-small` | OpenAI embedding model |
| `TEI_BASE_URL` | `http://localhost:8080` | TEI server URL |
| `TEI_EMBED_MODEL` | `BAAI/bge-small-en-v1.5` | TEI embedding model |
| `DEFAULT_LLM_PROVIDER` | — | LLM provider for RAG: `openai` \| `ollama` \| `anthropic` |
| `OPENAI_LLM_MODEL` | `gpt-4o-mini` | OpenAI LLM model |
| `OLLAMA_LLM_MODEL` | `llama3.2` | Ollama LLM model |
| `ANTHROPIC_API_KEY` | — | Anthropic API key |
| `ANTHROPIC_MODEL` | `claude-haiku-4-5-20251001` | Anthropic model |
| `WORKER_COUNT` | `5` | Concurrent ingestion workers |
| `EMBED_RATE_LIMIT` | `10` | Max embedding API calls per second |
| `MAX_UPLOAD_SIZE_MB` | `20` | Max file upload size in MB |
| `CHUNK_STRATEGY` | `auto` | Default chunking strategy |
| `CHUNK_SIZE` | `2000` | Max characters per chunk |
| `CHUNK_OVERLAP` | `200` | Overlap characters between chunks |
| `DEFAULT_PROMPT_SLUG` | `basic_rag` | Default RAG prompt template |
