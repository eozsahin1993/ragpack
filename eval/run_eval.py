#!/usr/bin/env python3
"""CLI entrypoint: dev-only eval harness over SciFact (IR) and SQuAD (RAGAS) tracks."""

import argparse
import os
import sys

from download import EVAL_DIR
from ragpack_client import RagPackClient
from results import load_baseline, print_aggregate_table, save_baseline, save_run

import scifact_track
import squad_track


def track_config(args):
    return {
        "embed_model": args.embed_model,
        "chunk_strategy": args.chunk_strategy,
        "chunk_size": args.chunk_size,
        "chunk_overlap": args.chunk_overlap,
        "seed": args.seed,
    }


def build_arg_parser():
    p = argparse.ArgumentParser(description="RagPack retrieval/generation quality eval harness")
    p.add_argument("--api-key", default=os.environ.get("RAGPACK_API_KEY"))
    p.add_argument("--base-url", default="http://localhost:9000/api/v1")
    p.add_argument("--track", choices=["scifact", "squad", "both"], default="both")

    p.add_argument("--embed-model", default=None)
    p.add_argument("--chunk-strategy", default=None)
    p.add_argument("--chunk-size", type=int, default=None)
    p.add_argument("--chunk-overlap", type=int, default=None)

    p.add_argument("--seed", type=int, default=42)
    p.add_argument("--ir-queries", type=int, default=50)
    p.add_argument("--ir-corpus-size", type=int, default=300, help="0 = full SciFact corpus")

    p.add_argument("--gen-questions", type=int, default=30)
    p.add_argument("--prompt-slug", default="basic_rag")
    p.add_argument("--model", default=None, help="LLM model passed to /rag (required for --track squad/both)")
    p.add_argument("--rag-top-k", type=int, default=5)
    p.add_argument("--judge-model", default="gpt-4o-mini", help="OpenAI model RAGAS uses as its judge/embedder")

    p.add_argument("--save-baseline", action="store_true")
    p.add_argument("--keep-collections", action="store_true", help="skip deleting eval collections after the run")
    return p


def main():
    try:
        from dotenv import load_dotenv
        load_dotenv(EVAL_DIR.parent / ".env")  # picks up OPENAI_API_KEY if it's set there
    except ImportError:
        pass

    args = build_arg_parser().parse_args()
    if not args.api_key:
        sys.exit("error: --api-key or RAGPACK_API_KEY is required")
    if args.track in ("squad", "both") and not args.model:
        sys.exit("error: --model is required for the squad track (passed to /rag)")
    if args.track in ("squad", "both") and not os.environ.get("OPENAI_API_KEY"):
        sys.exit("error: OPENAI_API_KEY must be set (env var or in repo-root .env) for the squad track (RAGAS judge)")

    client = RagPackClient(args.base_url, args.api_key)
    baseline = load_baseline()
    config = track_config(args)

    all_results = []
    if args.track in ("scifact", "both"):
        all_results.append(scifact_track.run(client, args, config))
    if args.track in ("squad", "both"):
        all_results.append(squad_track.run(client, args, config))

    for track_result in all_results:
        print_aggregate_table(track_result, baseline)

    save_run(all_results)
    if args.save_baseline:
        save_baseline(all_results)


if __name__ == "__main__":
    main()
