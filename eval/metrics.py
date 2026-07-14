"""Dependency-free IR metrics (nDCG@k, Recall@k) scored against BEIR-style qrels."""

import math


def dcg_at_k(relevances, k):
    return sum(rel / math.log2(idx + 2) for idx, rel in enumerate(relevances[:k]))


def ndcg_at_k(ranked_doc_ids, qrel, k):
    relevances = [qrel.get(d, 0) for d in ranked_doc_ids]
    dcg = dcg_at_k(relevances, k)
    idcg = dcg_at_k(sorted(qrel.values(), reverse=True), k)
    return dcg / idcg if idcg > 0 else 0.0


def recall_at_k(ranked_doc_ids, qrel, k):
    relevant = {d for d, score in qrel.items() if score > 0}
    if not relevant:
        return None
    retrieved = set(ranked_doc_ids[:k])
    return len(retrieved & relevant) / len(relevant)
