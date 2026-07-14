# Eval harness

Dev-only tool — not a product feature. Answers "did this change to chunking/embedding/prompt config help or hurt retrieval or generation quality?" by running RagPack's real API against two public, pre-existing benchmarks (not RagPack's own docs), so scores reflect the pipeline's general quality rather than a golden set someone wrote by hand.

- **`scifact`** — [BEIR](https://github.com/beir-cellar/beir)'s SciFact dataset, retrieval-only. Scored with `nDCG@10` / `Recall@100`. No LLM judge, no extra API key.
- **`squad`** — [SQuAD 2.0](https://rajpurkar.github.io/SQuAD-explorer/)'s dev set, generation quality via [RAGAS](https://github.com/explodinggraphs/ragas) (`faithfulness`, `answer_relevancy`, `context_precision`, `context_recall`). Needs `OPENAI_API_KEY` (used as the RAGAS judge) and a configured LLM provider on the RagPack stack itself.

Both tracks sample a small subset of their dataset (not the full corpus) so a run takes minutes, not hours — see `--ir-queries`/`--ir-corpus-size`/`--gen-questions` below.

## Setup

```bash
# from repo root
npm run dev              # or: npm run dev:ollama / npm run dev:tei — squad track needs *some* LLM provider configured
cd eval
python3 -m venv .venv && source .venv/bin/activate
pip install -r requirements.txt
```

Grab a RagPack API key (printed on first boot, or `ragpack logs backend | grep "Key:"` / `backend/.data/api_key`).

`OPENAI_API_KEY` (only needed for `--track squad`) is picked up automatically from the repo-root `.env` if present, or set it in your shell.

## Running

```bash
# retrieval-only, fast, free
python3 run_eval.py --api-key <key> --track scifact

# generation quality — needs --model (passed to /rag) and OPENAI_API_KEY
python3 run_eval.py --api-key <key> --track squad --model gpt-4o-mini

# both (default track is "both")
python3 run_eval.py --api-key <key> --model gpt-4o-mini
```

Every run prints a scored table and writes the full result to `results/<timestamp>.json`. If `results/baseline.json` exists, scores are also shown as a delta against it.

```bash
# after a run you're happy with — commits this as "the" baseline going forward
python3 run_eval.py --api-key <key> --model gpt-4o-mini --save-baseline
```

`results/baseline.json` is the only file under `results/` that's checked into git — it's the shared point of comparison, not a personal artifact.

## Flags

| Flag | Default | Applies to |
|---|---|---|
| `--api-key` | `$RAGPACK_API_KEY` | both |
| `--base-url` | `http://localhost:9000/api/v1` | both |
| `--track` | `both` | `scifact` \| `squad` \| `both` |
| `--embed-model` / `--chunk-strategy` / `--chunk-size` / `--chunk-overlap` | RagPack's own defaults | both — the actual levers you're testing |
| `--seed` | `42` | both — sampling reproducibility |
| `--ir-queries` | `50` | scifact |
| `--ir-corpus-size` | `300` (`0` = full 5,183-doc corpus) | scifact |
| `--gen-questions` | `30` | squad |
| `--prompt-slug` | `basic_rag` | squad |
| `--model` | *(required for squad)* | squad — LLM passed to `/rag` |
| `--rag-top-k` | `5` | squad |
| `--judge-model` | `gpt-4o-mini` | squad — RAGAS's judge/embedder |
| `--save-baseline` | off | both |
| `--keep-collections` | off | both — skip deleting the eval collections after the run, for debugging |

## Notes

- Each run creates a fresh, timestamped collection per track and deletes it when the run finishes (unless `--keep-collections`).
- Downloaded datasets are cached under `.cache/` (gitignored) — only downloaded once.
- SciFact's default 300-doc sample makes scores *not* directly comparable to published full-corpus BEIR leaderboard numbers (a smaller candidate pool is an easier retrieval task) — treat it as a self-comparison baseline, not a leaderboard entry, unless you run with `--ir-corpus-size 0`.
