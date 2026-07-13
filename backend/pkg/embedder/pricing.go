package embedder

// PerMTok is USD per 1M tokens. Add an entry here whenever a new priced
// embedding model is wired into the registry — models absent from this
// table are treated as unpriced, not free (see Cost).
var pricingTable = map[string]float64{
	"text-embedding-3-small": 0.02,
	"text-embedding-3-large": 0.13,
}

// Cost returns the USD cost of u for model, and whether model has a known
// price. Absence from pricingTable means "unpriced" (unknown model), which
// callers must not conflate with a confirmed $0 local-model cost — local
// providers (Ollama, self-hosted TEI) are tracked separately via
// Registry.IsLocal, since their model names are arbitrary and can't be
// recognized by lookup.
func Cost(model string, u Usage) (usd float64, priced bool) {
	perMTok, ok := pricingTable[model]
	if !ok {
		return 0, false
	}
	return float64(u.TotalTokens) / 1_000_000 * perMTok, true
}
