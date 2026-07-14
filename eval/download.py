"""Shared download-with-cache helper for the eval datasets."""

from pathlib import Path

EVAL_DIR = Path(__file__).parent
CACHE_DIR = EVAL_DIR / ".cache"
RESULTS_DIR = EVAL_DIR / "results"


def download(url, dest_path):
    import requests

    print(f"downloading {url} ...")
    dest_path.parent.mkdir(parents=True, exist_ok=True)
    with requests.get(url, stream=True, timeout=120) as resp:
        resp.raise_for_status()
        with open(dest_path, "wb") as f:
            for chunk in resp.iter_content(chunk_size=1 << 16):
                f.write(chunk)
