// Package ai defines the pluggable AI provider interface and a factory.
// New providers (OpenAI, Claude, DeepSeek, Groq, Ollama) can be registered
// without touching any business logic.
package ai

import (
	"context"

	"github.com/priyanjul/ai-finance-tracker/internal/dto"
)

// Provider is the AI backend abstraction.  Every AI integration must implement
// this interface so the rest of the codebase stays provider-agnostic.
type Provider interface {
	// ParseExpense extracts structured expense data from free text or an image.
	ParseExpense(ctx context.Context, text, imageURL string) (*dto.AIExpenseParseResponse, error)

	// CategorizeExpense classifies a merchant+description into a category.
	CategorizeExpense(ctx context.Context, merchant, description string) (string, error)

	// GenerateSummary produces a human-readable summary (weekly / monthly).
	GenerateSummary(ctx context.Context, data, summaryType string) (string, error)

	// GenerateInsights returns actionable personal finance tips.
	GenerateInsights(ctx context.Context, data map[string]interface{}) ([]string, error)

	// NLToSQLFilter converts natural-language search into a safe SQL WHERE clause.
	NLToSQLFilter(ctx context.Context, query, userID string) (string, error)

	// PredictExpenses forecasts end-of-month spend, savings, and budget risk.
	PredictExpenses(ctx context.Context, data map[string]interface{}) (*dto.PredictionData, error)

	// CalcHealthScore returns a 0-100 financial health score.
	CalcHealthScore(ctx context.Context, data map[string]interface{}) (float64, error)

	// ParseExpenseFromAudio transcribes audio and extracts structured expense
	// data in a single Gemini call (multimodal: audio + prompt).
	ParseExpenseFromAudio(ctx context.Context, audioData []byte, mimeType string) (*dto.AIVoiceParseResponse, error)
}

// ─── Factory ──────────────────────────────────────────────────────────────────

// Factory holds registered Provider implementations keyed by name.
type Factory struct {
	providers map[string]Provider
}

// NewFactory returns an empty Factory.
func NewFactory() *Factory {
	return &Factory{providers: make(map[string]Provider)}
}

// Register adds a provider under the given name (e.g. "gemini", "openai").
func (f *Factory) Register(name string, p Provider) {
	f.providers[name] = p
}

// Get returns the Provider registered under name, or nil if absent.
func (f *Factory) Get(name string) Provider {
	return f.providers[name]
}

// Default returns the first registered provider; useful when only one is set.
func (f *Factory) Default() Provider {
	for _, p := range f.providers {
		return p
	}
	return nil
}

// List returns all registered provider names.
func (f *Factory) List() []string {
	names := make([]string, 0, len(f.providers))
	for n := range f.providers {
		names = append(names, n)
	}
	return names
}
