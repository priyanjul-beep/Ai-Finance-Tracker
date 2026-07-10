// Package ai – Gemini provider implementation.
package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"

	"github.com/priyanjul/ai-finance-tracker/internal/dto"
)

// GeminiProvider implements Provider using Google's Gemini API.
type GeminiProvider struct {
	client *genai.Client
	model  string
}

// NewGeminiProvider creates an authenticated Gemini client.
// model defaults to "gemini-2.0-flash" when empty.
func NewGeminiProvider(apiKey, model string) (*GeminiProvider, error) {
	if model == "" {
		model = "gemini-2.0-flash"
	}
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("gemini: new client: %w", err)
	}
	return &GeminiProvider{client: client, model: model}, nil
}

// Close releases the Gemini client resources.
func (g *GeminiProvider) Close() { g.client.Close() }

// ─── Provider implementation ──────────────────────────────────────────────────

// ParseExpense extracts amount, merchant, category, date, etc. from text.
func (g *GeminiProvider) ParseExpense(ctx context.Context, text, _ string) (*dto.AIExpenseParseResponse, error) {
	prompt := fmt.Sprintf(`
You are a financial data extractor. Extract expense details from the user input below.

Input: "%s"

Return ONLY valid JSON (no markdown) with these keys:
{
  "amount": <number>,
  "merchant": "<string>",
  "category": "<one of: food|travel|shopping|entertainment|health|investment|education|bills|recharge|fuel|rent|salary|utilities|subscription|personal_care|gift|charitable|insurance|others|unknown>",
  "date": "<ISO-8601 or null>",
  "notes": "<string>",
  "expense_type": "<spend|refund|transfer>",
  "payment_method": "<cash|card|upi|bank|wallet|online|unknown>",
  "confidence": <0.0-1.0>
}`, text)

	raw, err := g.generate(ctx, prompt)
	if err != nil {
		return nil, err
	}

	var result dto.AIExpenseParseResponse
	if err := json.Unmarshal([]byte(cleanJSON(raw)), &result); err != nil {
		return nil, fmt.Errorf("gemini: parse expense response: %w", err)
	}
	return &result, nil
}

// CategorizeExpense classifies a merchant+description pair.
func (g *GeminiProvider) CategorizeExpense(ctx context.Context, merchant, description string) (string, error) {
	prompt := fmt.Sprintf(`
Classify the following expense into exactly one category.
Merchant: %s
Description: %s

Reply with ONLY the category name from this list:
food, travel, shopping, entertainment, health, investment, education, bills,
recharge, fuel, rent, salary, utilities, subscription, personal_care, gift,
charitable, insurance, others, unknown`, merchant, description)

	raw, err := g.generate(ctx, prompt)
	if err != nil {
		return "unknown", err
	}
	return strings.TrimSpace(strings.ToLower(raw)), nil
}

// GenerateSummary produces a weekly or monthly plain-text summary.
func (g *GeminiProvider) GenerateSummary(ctx context.Context, data, summaryType string) (string, error) {
	var instruction string
	switch summaryType {
	case "weekly":
		instruction = "Write a concise weekly financial summary (max 150 words). Highlight top spending categories and give one saving tip."
	default:
		instruction = "Write a detailed monthly financial report (max 350 words). Include income, expenses, savings rate, top categories, and 3 actionable recommendations."
	}

	prompt := fmt.Sprintf("%s\n\nFinancial data:\n%s", instruction, data)
	return g.generate(ctx, prompt)
}

// GenerateInsights returns 5-7 actionable financial insights as a JSON array.
func (g *GeminiProvider) GenerateInsights(ctx context.Context, data map[string]interface{}) ([]string, error) {
	jsonData, _ := json.Marshal(data)
	prompt := fmt.Sprintf(`
Analyse the following personal finance data and return 5-7 actionable insights.

Data: %s

Return ONLY a JSON array of strings, e.g.:
["You spent 35%% more on food this month.", "Consider cutting 2 streaming subscriptions."]`, string(jsonData))

	raw, err := g.generate(ctx, prompt)
	if err != nil {
		return nil, err
	}

	var insights []string
	if err := json.Unmarshal([]byte(cleanJSON(raw)), &insights); err != nil {
		return nil, fmt.Errorf("gemini: parse insights: %w", err)
	}
	return insights, nil
}

// NLToSQLFilter converts a natural language query to a safe PostgreSQL WHERE clause.
func (g *GeminiProvider) NLToSQLFilter(ctx context.Context, query, userID string) (string, error) {
	prompt := fmt.Sprintf(`
Convert the following natural-language expense search query into a safe PostgreSQL WHERE clause.
Available columns: amount (numeric), category (text), merchant (text), date (timestamptz), expense_type (text), payment_method (text).
The clause must NOT contain subqueries, UNION, DROP, DELETE, UPDATE, INSERT, or any DDL.
Parameterise user_id as the literal string '%s' (already safe).

Query: "%s"

Return ONLY the WHERE clause, starting with the conditions (no "WHERE" keyword).
Example output: amount > 1000 AND category = 'food'`, userID, query)

	return g.generate(ctx, prompt)
}

// PredictExpenses forecasts end-of-month figures from historical data.
func (g *GeminiProvider) PredictExpenses(ctx context.Context, data map[string]interface{}) (*dto.PredictionData, error) {
	jsonData, _ := json.Marshal(data)
	prompt := fmt.Sprintf(`
Based on the following historical expense data, predict end-of-month figures.

Data: %s

Return ONLY valid JSON with these keys:
{
  "end_of_month_spending": <number>,
  "expected_savings": <number>,
  "budget_overrun_risk": <0-100>,
  "savings_goal_on_track": <true|false>
}`, string(jsonData))

	raw, err := g.generate(ctx, prompt)
	if err != nil {
		return nil, err
	}

	var result dto.PredictionData
	if err := json.Unmarshal([]byte(cleanJSON(raw)), &result); err != nil {
		return nil, fmt.Errorf("gemini: parse predictions: %w", err)
	}
	return &result, nil
}

// CalcHealthScore returns a 0-100 financial health score.
func (g *GeminiProvider) CalcHealthScore(ctx context.Context, data map[string]interface{}) (float64, error) {
	jsonData, _ := json.Marshal(data)
	prompt := fmt.Sprintf(`
Given the following financial profile, compute a health score from 0 to 100.
Consider: income stability, savings rate, expense ratio, budget adherence, subscriptions, and debt.

Data: %s

Return ONLY valid JSON: {"score": <0-100>}`, string(jsonData))

	raw, err := g.generate(ctx, prompt)
	if err != nil {
		return 0, err
	}

	var result struct {
		Score float64 `json:"score"`
	}
	if err := json.Unmarshal([]byte(cleanJSON(raw)), &result); err != nil {
		return 0, fmt.Errorf("gemini: parse health score: %w", err)
	}
	return result.Score, nil
}

// ParseExpenseFromImage sends an image to Gemini Vision for receipt/screenshot
// analysis. ocrText is optional pre-extracted text that improves accuracy.
func (g *GeminiProvider) ParseExpenseFromImage(ctx context.Context, imageData []byte, mimeType, ocrText string) (*dto.AIReceiptScanResponse, error) {
	if len(imageData) == 0 {
		return nil, fmt.Errorf("gemini: empty image data")
	}

	promptText := `You are an expert financial document parser. Examine this image carefully — it may be a UPI payment screenshot, PhonePe/GPay/Paytm confirmation, bank transfer notification, restaurant bill, shopping receipt, invoice, grocery bill, utility bill, fuel receipt, or any payment proof.

Extract ALL financial details and return ONLY valid JSON with no markdown fences:

{
  "amount": <total amount paid — number>,
  "merchant": "<merchant/store/payee name as shown on the receipt>",
  "category": "<one of: food|travel|shopping|entertainment|health|investment|education|bills|recharge|fuel|rent|salary|utilities|subscription|personal_care|gift|others|unknown>",
  "payment_method": "<one of: cash|card|upi|bank|wallet|online>",
  "expense_type": "<spend|refund|transfer — use 'refund' for cashback/reversal>",
  "currency": "<3-letter ISO code: INR|USD|EUR|GBP — default INR if unclear>",
  "date": "<YYYY-MM-DD>",
  "transaction_id": "<UPI Ref No / Order ID / Transaction ID or empty string>",
  "notes": "<brief purchase description or empty string>",
  "tax_amount": <GST/VAT/tax as number or 0>,
  "invoice_number": "<invoice or bill number or empty string>",
  "confidence": <0.0-1.0>
}

Category rules:
• Swiggy|Zomato|Domino's|McDonald's|KFC|restaurant|food delivery → food
• Uber|Ola|Rapido|IRCTC|bus|metro|flight|MakeMyTrip → travel
• Amazon|Flipkart|Myntra|Meesho|Ajio|Nykaa|retail shop → shopping
• Blinkit|Zepto|BigBasket|Dunzo|grocery|supermarket → food
• Apollo|Medplus|1mg|PharmEasy|hospital|clinic|doctor|pharmacy → health
• BESCOM|CESC|Tata Power|electricity|water board|gas bill → bills
• Airtel|Jio|Vi|BSNL|recharge|mobile bill|internet → recharge
• Indian Oil|HPCL|BPCL|Shell|HP|petrol pump|fuel → fuel
• Netflix|Spotify|Amazon Prime|Disney+|YouTube Premium → subscription
• college|school|fees|tuition|course|university → education
• Reliance Smart|D-Mart|More|Spencer's|supermarket → shopping

Payment method hints:
• "UPI"|"UPI Ref"|"VPA"|"GPay"|"PhonePe"|"Paytm" → upi
• "Credit Card"|"Debit Card"|"VISA"|"Mastercard"|"RuPay" → card
• "NEFT"|"RTGS"|"IMPS"|"Net Banking" → bank
• "Cash"|"COD" → cash
• "Wallet"|"Mobikwik"|"Ola Money" → wallet`

	var parts []genai.Part
	parts = append(parts, genai.Text(promptText))
	parts = append(parts, genai.Blob{MIMEType: mimeType, Data: imageData})

	if ocrText != "" {
		parts = append(parts, genai.Text(fmt.Sprintf(
			"\n\nAdditional OCR text extracted from the image (use as supplementary hint):\n---\n%s\n---",
			ocrText,
		)))
	}

	model := g.client.GenerativeModel(g.model)
	resp, err := model.GenerateContent(ctx, parts...)
	if err != nil {
		return nil, fmt.Errorf("gemini: receipt scan: %w", err)
	}
	if len(resp.Candidates) == 0 {
		return nil, fmt.Errorf("gemini: receipt scan: no candidates")
	}

	var sb strings.Builder
	for _, part := range resp.Candidates[0].Content.Parts {
		sb.WriteString(fmt.Sprintf("%v", part))
	}
	raw := cleanJSON(sb.String())

	var result dto.AIReceiptScanResponse
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		return nil, fmt.Errorf("gemini: receipt scan unmarshal: %w (raw=%s)", err, raw)
	}
	return &result, nil
}

// ParseExpenseFromAudio sends audio bytes to Gemini (multimodal) and returns
// a fully parsed expense including the transcript. Gemini 1.5-flash can
// process audio inline up to ~9.5 MB without uploading to Files API.
func (g *GeminiProvider) ParseExpenseFromAudio(ctx context.Context, audioData []byte, mimeType string) (*dto.AIVoiceParseResponse, error) {
	if len(audioData) == 0 {
		return nil, fmt.Errorf("gemini: empty audio data")
	}

	prompt := genai.Text(`You are an expert financial assistant. Listen to this audio and extract expense details.

Return ONLY valid JSON (no markdown fences, no explanation) with these exact keys:
{
  "transcript": "<verbatim words spoken>",
  "amount": <number or null>,
  "merchant": "<merchant/vendor/payee name or null>",
  "category": "<one of: food|travel|shopping|entertainment|health|investment|education|bills|recharge|fuel|rent|salary|utilities|subscription|personal_care|gift|others|unknown or null>",
  "date": "<today|yesterday|YYYY-MM-DD — use 'today' if not mentioned>",
  "notes": "<any extra context, empty string if none>",
  "expense_type": "<spend|refund|transfer — default 'spend'>",
  "payment_method": "<cash|card|upi|bank|wallet|online or null>",
  "confidence": <0.0-1.0>
}

Inference rules:
- "Swiggy", "Zomato", "restaurant", "lunch", "dinner" → food
- "Uber", "Ola", "petrol", "fuel", "auto", "bus", "metro" → travel/fuel
- "Amazon", "Flipkart", "shoes", "clothes", "shopping" → shopping
- "Netflix", "Spotify", "subscription" → subscription
- "doctor", "hospital", "medicine", "pharmacy" → health
- "electricity", "water bill", "internet", "phone bill" → bills
- "rent", "house", "flat" → rent
- "salary", "credited" → salary
- "college", "school", "fees", "course" → education
- Return null only for fields that cannot be inferred at all`)

	audioPart := genai.Blob{
		MIMEType: mimeType,
		Data:     audioData,
	}

	model := g.client.GenerativeModel(g.model)
	resp, err := model.GenerateContent(ctx, prompt, audioPart)
	if err != nil {
		return nil, fmt.Errorf("gemini: voice parse generate: %w", err)
	}
	if len(resp.Candidates) == 0 {
		return nil, fmt.Errorf("gemini: voice parse: no candidates")
	}

	var sb strings.Builder
	for _, part := range resp.Candidates[0].Content.Parts {
		sb.WriteString(fmt.Sprintf("%v", part))
	}
	raw := cleanJSON(sb.String())

	var result dto.AIVoiceParseResponse
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		return nil, fmt.Errorf("gemini: voice parse unmarshal: %w (raw=%s)", err, raw)
	}
	return &result, nil
}

// ─── internal helpers ─────────────────────────────────────────────────────────

// generate calls Gemini and returns the first candidate's text.
func (g *GeminiProvider) generate(ctx context.Context, prompt string) (string, error) {
	model := g.client.GenerativeModel(g.model)
	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", fmt.Errorf("gemini: generate content: %w", err)
	}
	if len(resp.Candidates) == 0 {
		return "", fmt.Errorf("gemini: no candidates returned")
	}
	var sb strings.Builder
	for _, part := range resp.Candidates[0].Content.Parts {
		sb.WriteString(fmt.Sprintf("%v", part))
	}
	return sb.String(), nil
}

// cleanJSON strips markdown code fences that Gemini sometimes wraps around JSON.
func cleanJSON(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "```json")
	s = strings.TrimPrefix(s, "```")
	s = strings.TrimSuffix(s, "```")
	return strings.TrimSpace(s)
}
