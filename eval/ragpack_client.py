"""Thin HTTP client for the RagPack public API, plus an ingest-completion poller."""

import sys
import time


class RagPackClient:
    def __init__(self, base_url, api_key):
        import requests

        self.base_url = base_url.rstrip("/")
        self.session = requests.Session()
        self.session.headers["Authorization"] = f"Bearer {api_key}"

    def _check(self, resp):
        if not resp.ok:
            sys.exit(f"error: {resp.request.method} {resp.request.url} -> {resp.status_code}: {resp.text}")
        return resp

    def create_collection(self, name, embed_model=None, chunk_strategy=None, chunk_size=None, chunk_overlap=None):
        body = {"name": name}
        if embed_model:
            body["embed_model"] = embed_model
        chunk_config = {}
        if chunk_strategy:
            chunk_config["strategy"] = chunk_strategy
        if chunk_size:
            chunk_config["size"] = chunk_size
        if chunk_overlap is not None:
            chunk_config["overlap"] = chunk_overlap
        if chunk_config:
            body["chunk_config"] = chunk_config
        resp = self._check(self.session.post(f"{self.base_url}/collections", json=body))
        return resp.json()

    def delete_collection(self, slug):
        self._check(self.session.delete(f"{self.base_url}/collections/{slug}"))

    def ingest_file(self, slug, filename, content):
        files = {"file": (filename, content, "text/plain")}
        self._check(self.session.post(f"{self.base_url}/collections/{slug}/ingest", files=files))

    def list_documents(self, slug, limit=500):
        resp = self._check(self.session.get(f"{self.base_url}/collections/{slug}/documents", params={"limit": limit}))
        return resp.json()

    def query(self, slug, query_text, top_k):
        body = {"query": query_text, "top_k": top_k}
        resp = self._check(self.session.post(f"{self.base_url}/collections/{slug}/query", json=body))
        return resp.json()

    def rag(self, slug, query_text, prompt_slug, model, top_k):
        body = {"query": query_text, "prompt_slug": prompt_slug, "model": model, "top_k": top_k}
        resp = self._check(self.session.post(f"{self.base_url}/collections/{slug}/rag", json=body))
        return resp.json()


def wait_for_completion(client, slug, expected_count, timeout=900, poll_interval=3):
    deadline = time.time() + timeout
    while time.time() < deadline:
        docs = client.list_documents(slug, limit=max(expected_count, 1))["documents"] or []
        if len(docs) >= expected_count and all(d["status"] != "ingesting" for d in docs):
            failed = [d for d in docs if d["status"] == "failed"]
            if failed:
                print(f"warning: {len(failed)} document(s) failed to ingest", file=sys.stderr)
                for d in failed:
                    print(f"  {d['file_uri']}: {d.get('error')}", file=sys.stderr)
            return docs
        time.sleep(poll_interval)
    raise TimeoutError(f"ingestion did not complete within {timeout}s")
