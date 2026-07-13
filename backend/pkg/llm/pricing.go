package llm

import "regexp"

// Pricing is USD per 1M tokens. Add an entry here whenever a new priced
// model is wired into the registry — models absent from this table are
// treated as unpriced, not free (see Cost).
type Pricing struct {
	InputPerMTok  float64
	OutputPerMTok float64
}

var pricingTable = map[string]Pricing{
	"gpt-4o":      {InputPerMTok: 2.50, OutputPerMTok: 10.00},
	"gpt-4o-mini": {InputPerMTok: 0.15, OutputPerMTok: 0.60},

	// Sonnet 5 has intro pricing ($2/$10) through 2026-08-31; sticker rate used here.
	"claude-opus-4-8":  {InputPerMTok: 5.00, OutputPerMTok: 25.00},
	"claude-sonnet-5":  {InputPerMTok: 3.00, OutputPerMTok: 15.00},
	"claude-haiku-4-5": {InputPerMTok: 1.00, OutputPerMTok: 5.00},
}

// Anthropic models are often configured with dated IDs (claude-haiku-4-5-20251001);
// the table keys on the alias, so strip a trailing -YYYYMMDD before lookup.
var dateSuffix = regexp.MustCompile(`-\d{8}$`)

// Cost returns the USD cost of u for model, and whether model has a known
// price. Absence from pricingTable means "unpriced" (unknown model),
// which callers must not conflate with a confirmed $0 local-model cost —
// local providers are tracked separately via Registry.IsLocal, since an
// Ollama model's name is arbitrary and can't be recognized by lookup.
func Cost(model string, u Usage) (usd float64, priced bool) {
	p, ok := pricingTable[model]
	if !ok {
		p, ok = pricingTable[dateSuffix.ReplaceAllString(model, "")]
	}
	if !ok {
		return 0, false
	}
	usd = float64(u.InputTokens)/1_000_000*p.InputPerMTok + float64(u.OutputTokens)/1_000_000*p.OutputPerMTok
	return usd, true
}
