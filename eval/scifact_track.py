"""Track 1: SciFact (BEIR) retrieval-only IR benchmark - nDCG@10 / Recall@100, no LLM judge."""

import json
import random
import re
import time
import zipfile
from statistics import mean

from download import CACHE_DIR, download
from metrics import ndcg_at_k, recall_at_k
from ragpack_client import wait_for_completion

SCIFACT_URL = "https://public.ukp.informatik.tu-darmstadt.de/thakur/BEIR/datasets/scifact.zip"
FILE_URI_RE = re.compile(r"^upload://scifact_(.+)\.txt$")


def load_scifact():
    extract_dir = CACHE_DIR / "scifact"
    if not (extract_dir / "corpus.jsonl").exists():
        zip_path = CACHE_DIR / "scifact.zip"
        download(SCIFACT_URL, zip_path)
        with zipfile.ZipFile(zip_path) as zf:
            zf.extractall(CACHE_DIR)
        zip_path.unlink()

    corpus = {}
    with open(extract_dir / "corpus.jsonl") as f:
        for line in f:
            doc = json.loads(line)
            corpus[doc["_id"]] = doc

    queries = {}
    with open(extract_dir / "queries.jsonl") as f:
        for line in f:
            q = json.loads(line)
            queries[q["_id"]] = q["text"]

    qrels = {}
    with open(extract_dir / "qrels" / "test.tsv") as f:
        next(f)  # header row
        for line in f:
            qid, docid, score = line.strip().split("\t")
            qrels.setdefault(qid, {})[docid] = int(score)

    return corpus, queries, qrels


def sample_scifact(corpus, queries, qrels, seed, n_queries, corpus_size):
    rng = random.Random(seed)
    candidate_qids = [qid for qid in queries if qrels.get(qid)]
    rng.shuffle(candidate_qids)
    sampled_qids = candidate_qids[:n_queries]

    relevant_doc_ids = set()
    for qid in sampled_qids:
        relevant_doc_ids.update(qrels[qid].keys())

    sampled_docs = {docid: corpus[docid] for docid in relevant_doc_ids if docid in corpus}

    if corpus_size and len(sampled_docs) < corpus_size:
        remaining = [d for d in corpus if d not in sampled_docs]
        rng.shuffle(remaining)
        for docid in remaining[: corpus_size - len(sampled_docs)]:
            sampled_docs[docid] = corpus[docid]

    sampled_queries = {qid: queries[qid] for qid in sampled_qids}
    sampled_qrels = {qid: qrels[qid] for qid in sampled_qids}
    return sampled_docs, sampled_queries, sampled_qrels


def run(client, args, track_config):
    print("\n== Track: scifact (IR) ==")
    corpus, queries, qrels = load_scifact()
    docs, sampled_queries, sampled_qrels = sample_scifact(
        corpus, queries, qrels, args.seed, args.ir_queries, args.ir_corpus_size
    )
    print(f"sampled {len(docs)} docs, {len(sampled_queries)} queries")

    collection = client.create_collection(
        name=f"eval-scifact-{int(time.time())}",
        embed_model=args.embed_model,
        chunk_strategy=args.chunk_strategy,
        chunk_size=args.chunk_size,
        chunk_overlap=args.chunk_overlap,
    )
    slug = collection["slug"]
    print(f"created collection {slug}")

    try:
        for docid, doc in docs.items():
            content = f"{doc.get('title', '')}\n\n{doc.get('text', '')}".encode("utf-8")
            client.ingest_file(slug, f"scifact_{docid}.txt", content)

        wait_for_completion(client, slug, expected_count=len(docs))

        per_query = []
        for qid, qtext in sampled_queries.items():
            result = client.query(slug, qtext, top_k=100)
            ranked_doc_ids, seen = [], set()
            for item in result.get("results", []):
                m = FILE_URI_RE.match(item.get("file_uri") or "")
                if m and m.group(1) not in seen:
                    seen.add(m.group(1))
                    ranked_doc_ids.append(m.group(1))
            qrel = sampled_qrels[qid]
            per_query.append({
                "query_id": qid,
                "query": qtext,
                "ndcg@10": ndcg_at_k(ranked_doc_ids, qrel, 10),
                "recall@100": recall_at_k(ranked_doc_ids, qrel, 100),
            })

        aggregate = {
            "ndcg@10": mean(x["ndcg@10"] for x in per_query),
            "recall@100": mean(x["recall@100"] for x in per_query if x["recall@100"] is not None),
        }
        return {"track": "scifact", "config": track_config, "per_item": per_query, "aggregate": aggregate}
    finally:
        if not args.keep_collections:
            client.delete_collection(slug)
