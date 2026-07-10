package db

import "sort"

// MergeWeightedRRF fuses ranked vector and FTS result lists via weighted RRF:
// per channel, weight / (rrf_k + rank), summed across channels a result
// appears in. Backend-agnostic since lancedb's built-in reranker has no
// per-channel weight option.
func MergeWeightedRRF(vectorResults, ftsResults []ChunkQueryResult, topK int, s HybridSettings) []ChunkQueryResult {
	merged := make(map[string]*ChunkQueryResult, len(vectorResults)+len(ftsResults))
	rawScores := make(map[string]float32, len(vectorResults)+len(ftsResults))

	for rank, r := range vectorResults {
		rCopy := r
		merged[r.ID] = &rCopy
		rawScores[r.ID] = s.SemanticWeight / (s.RRFK + float32(rank))
	}
	for rank, r := range ftsResults {
		contribution := s.FullTextWeight / (s.RRFK + float32(rank))
		if existing, ok := merged[r.ID]; ok {
			existing.KeywordBM25Score = r.KeywordBM25Score
			rawScores[r.ID] += contribution
		} else {
			rCopy := r
			merged[r.ID] = &rCopy
			rawScores[r.ID] = contribution
		}
	}

	results := make([]ChunkQueryResult, 0, len(merged))
	for id, r := range merged {
		r.RRFScore = rawScores[id]
		results = append(results, *r)
	}
	// Map iteration order is randomized, so a tie needs an explicit tiebreaker
	// (chunk ID) or ranking becomes non-deterministic across identical calls.
	sort.Slice(results, func(i, j int) bool {
		if results[i].RRFScore != results[j].RRFScore {
			return results[i].RRFScore > results[j].RRFScore
		}
		return results[i].ID < results[j].ID
	})
	if len(results) > topK {
		results = results[:topK]
	}

	// Normalize against this batch's own top score, so the best result here
	// is always 100% — a fixed ceiling based on weights/k alone gets
	// exceeded routinely by real dual-channel matches (a chunk found in
	// both channels sums both contributions).
	if len(results) > 0 && results[0].RRFScore > 0 {
		ceiling := results[0].RRFScore
		for i := range results {
			results[i].RRFScoreNormalized = (results[i].RRFScore / ceiling) * 100
		}
	}

	return results
}
