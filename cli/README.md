# RagPack CLI

Self-hostable semantic search and RAG infrastructure — up and running in minutes.

## Install

```bash
npm install -g ragpack
```

Or use without installing:

```bash
npx ragpack <command>
```

## Quick start

```bash
ragpack init          # create .env.ragpack in the current directory
ragpack start         # start the stack (API :9000, admin UI :3000)
```

With Ollama (fully local embeddings):

```bash
ragpack start --profile ollama
```

## Commands

| Command | Description |
|---|---|
| `ragpack init` | Create `.env.ragpack` in the current directory |
| `ragpack start [--profile ollama]` | Start the stack |
| `ragpack stop [-v]` | Stop the stack (`-v` removes volumes and all data) |
| `ragpack logs [service]` | Tail logs (`backend`, `web-admin`, `ollama`) |
| `ragpack update` | Pull latest images and restart |
| `ragpack eject` | Copy `docker-compose.yml` locally for customization |

## Full documentation

See the [RagPack repository](https://github.com/eozsahin1993/ragpack) for configuration options, API reference, and embedding model support.

## License

MIT
