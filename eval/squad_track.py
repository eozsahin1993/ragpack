"""Track 2: SQuAD 2.0 generation-quality benchmark, scored with RAGAS (needs OPENAI_API_KEY)."""

import json
import math
import random
import re
import sys
import time
from statistics import mean

from download import CACHE_DIR, download
from ragpack_client import wait_for_completion

SQUAD_URL = "https://rajpurkar.github.io/SQuAD-explorer/dataset/dev-v2.0.json"

# $/token, judge-LLM chat calls only (embedding-call cost isn't captured by
# ragas's token_usage_parser, but text-embedding-3-small is negligible: $0.02/1M).
KNOWN_JUDGE_PRICING = {
    "gpt-4o-mini": (0.15 / 1_000_000, 0.60 / 1_000_000),  # (input, output)
}


def load_squad():
    path = CACHE_DIR / "squad" / "dev-v2.0.json"
    if not path.exists():
        download(SQUAD_URL, path)
    with open(path) as f:
        return json.load(f)["data"]


def slugify(text):
    return re.sub(r"[^a-zA-Z0-9]+", "_", text).strip("_").lower()


def sample_squad(articles, seed, n_questions):
    rng = random.Random(seed)
    candidates = []
    for article in articles:
        title = article["title"]
        for para_idx, para in enumerate(article["paragraphs"]):
            for qa in para["qas"]:
                if qa.get("is_impossible") or not qa["answers"]:
                    continue
                candidates.append({
                    "para_key": f"{slugify(title)}_{para_idx}",
                    "context": para["context"],
                    "question": qa["question"],
                    "answer": qa["answers"][0]["text"],
                })
    rng.shuffle(candidates)
    sampled = candidates[:n_questions]

    paragraphs = {item["para_key"]: item["context"] for item in sampled}
    return sampled, paragraphs


def run(client, args, track_config):
    print("\n== Track: squad (RAGAS) ==")
    try:
        from datasets import Dataset
        from langchain_openai import ChatOpenAI, OpenAIEmbeddings
        from ragas import evaluate as ragas_evaluate
        from ragas.cost import get_token_usage_for_openai
        from ragas.embeddings import LangchainEmbeddingsWrapper
        from ragas.llms import LangchainLLMWrapper
        from ragas.metrics import answer_relevancy, context_precision, context_recall, faithfulness
    except ImportError:
        sys.exit("error: the squad track needs `pip install -r eval/requirements.txt`")

    articles = load_squad()
    sampled, paragraphs = sample_squad(articles, args.seed, args.gen_questions)
    print(f"sampled {len(sampled)} questions across {len(paragraphs)} paragraphs")

    collection = client.create_collection(
        name=f"eval-squad-{int(time.time())}",
        embed_model=args.embed_model,
        chunk_strategy=args.chunk_strategy,
        chunk_size=args.chunk_size,
        chunk_overlap=args.chunk_overlap,
    )
    slug = collection["slug"]
    print(f"created collection {slug}")

    try:
        for key, context in paragraphs.items():
            client.ingest_file(slug, f"squad_{key}.txt", context.encode("utf-8"))

        wait_for_completion(client, slug, expected_count=len(paragraphs))

        questions, contexts_list, answers, ground_truths = [], [], [], []
        for item in sampled:
            resp = client.rag(slug, item["question"], args.prompt_slug, args.model, top_k=args.rag_top_k)
            questions.append(item["question"])
            contexts_list.append([c["chunk_text"] for c in resp.get("chunks", []) if c.get("chunk_text")])
            answers.append(resp.get("answer", ""))
            ground_truths.append(item["answer"])

        dataset = Dataset.from_dict({
            "question": questions,
            "contexts": contexts_list,
            "answer": answers,
            "ground_truth": ground_truths,
        })

        # ragas's non-deprecated LLM/embeddings APIs (llm_factory, embeddings.OpenAIEmbeddings)
        # don't implement embed_query() (needed by the legacy metric classes imported above)
        # or route through the callback ragas.cost uses for token/cost tracking - the
        # Langchain-wrapped classes are deprecated but are what actually works with both.
        judge_llm = LangchainLLMWrapper(ChatOpenAI(model=args.judge_model))
        judge_embeddings = LangchainEmbeddingsWrapper(OpenAIEmbeddings(model="text-embedding-3-small"))
        result = ragas_evaluate(
            dataset,
            metrics=[faithfulness, answer_relevancy, context_precision, context_recall],
            llm=judge_llm,
            embeddings=judge_embeddings,
            token_usage_parser=get_token_usage_for_openai,
        )
        scores = result.to_pandas()

        per_query = []
        for i, question in enumerate(questions):
            per_query.append({
                "question": question,
                "answer": answers[i],
                "ground_truth": ground_truths[i],
                "contexts": contexts_list[i],
                "faithfulness": float(scores.loc[i, "faithfulness"]),
                "answer_relevancy": float(scores.loc[i, "answer_relevancy"]),
                "context_precision": float(scores.loc[i, "context_precision"]),
                "context_recall": float(scores.loc[i, "context_recall"]),
            })

        aggregate = {}
        for metric in ("faithfulness", "answer_relevancy", "context_precision", "context_recall"):
            values = [x[metric] for x in per_query if not math.isnan(x[metric])]
            if len(values) < len(per_query):
                print(f"warning: {len(per_query) - len(values)}/{len(per_query)} '{metric}' scores were NaN", file=sys.stderr)
            aggregate[metric] = mean(values) if values else None

        cost = None
        pricing = KNOWN_JUDGE_PRICING.get(args.judge_model)
        if pricing:
            try:
                cost = {
                    "judge_model": args.judge_model,
                    "estimated_usd": result.total_cost(cost_per_input_token=pricing[0], cost_per_output_token=pricing[1]),
                    "note": "judge-LLM chat calls only; embedding-call cost (answer_relevancy) not included, negligible at $0.02/1M tokens",
                }
            except Exception as e:
                print(f"warning: could not compute judge cost: {e}", file=sys.stderr)

        return {
            "track": "squad",
            "config": track_config,
            "per_item": per_query,
            "aggregate": aggregate,
            "cost": cost,
        }
    finally:
        if not args.keep_collections:
            client.delete_collection(slug)
