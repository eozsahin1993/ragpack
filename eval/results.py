"""Run output: per-run JSON, committed baseline, and the stdout score/delta table."""

import json
import time

from download import RESULTS_DIR


def save_run(all_results):
    RESULTS_DIR.mkdir(parents=True, exist_ok=True)
    timestamp = time.strftime("%Y%m%dT%H%M%SZ", time.gmtime())
    path = RESULTS_DIR / f"{timestamp}.json"
    with open(path, "w") as f:
        json.dump(all_results, f, indent=2)
    print(f"\nsaved results to {path}")


def load_baseline():
    path = RESULTS_DIR / "baseline.json"
    if not path.exists():
        return {}
    with open(path) as f:
        return json.load(f)


def save_baseline(all_results):
    RESULTS_DIR.mkdir(parents=True, exist_ok=True)
    baseline = load_baseline()
    for track_result in all_results:
        baseline[track_result["track"]] = track_result["aggregate"]
    with open(RESULTS_DIR / "baseline.json", "w") as f:
        json.dump(baseline, f, indent=2)
    print(f"saved baseline to {RESULTS_DIR / 'baseline.json'}")


def print_aggregate_table(track_result, baseline):
    track = track_result["track"]
    prev = baseline.get(track)
    print(f"\n--- {track} aggregate ({len(track_result['per_item'])} items) ---")
    for metric, value in track_result["aggregate"].items():
        if value is None:
            print(f"  {metric:<20} n/a (no valid scores)")
            continue
        line = f"  {metric:<20} {value:.4f}"
        if prev and metric in prev and prev[metric] is not None:
            delta = value - prev[metric]
            line += f"   (baseline {prev[metric]:.4f}, Δ {delta:+.4f})"
        print(line)
    cost = track_result.get("cost")
    if cost:
        print(f"  judge cost (~{cost['judge_model']}): ${cost['estimated_usd']:.4f}  ({cost['note']})")
