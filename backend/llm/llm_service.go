package llm

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"backend/stt"
)

// LLMEvaluationResult strictly specifies the required Orchestrator schema
type LLMEvaluationResult struct {
	RiskBand       string   `json:"risk_band"`
	Flags          []string `json:"flags"`
	Recommendation string   `json:"recommendation"`
	Confidence     float64  `json:"confidence"`
}

// EvaluateRisk takes the STT transcription data and builds a comprehensive evaluation
func EvaluateRisk(sessionID string, sttData *stt.ExtractedData) *LLMEvaluationResult {
	log.Printf("[LLM-%s] Instantiating Orchestrator Prompt Synthesis from STT...", sessionID)

	// Combine into structured format matching LLM expected prompt engineering context
	prompt := buildSystemPrompt(sttData)
	
	// Simulate Network Request to LLM Framework (Claude / Gemini App)
	return mockExecuteLLMRequest(sessionID, prompt)
}

func buildSystemPrompt(s *stt.ExtractedData) string {
	return fmt.Sprintf(`
You are an advanced underwriting risk orchestrator determining fraud and viability for a live video loan onboarding module.
Evaluate the following STT transcription signals thoroughly.

=== STT Telemetry ===
Name: %s
Declared Income: %d
Loan Purpose: %s
Employment: %s
Verbal Consent: %v

=== INSTRUCTIONS ===
1. Analyze loan purpose intent and employment context.
2. Determine risk based strictly on transcription.
3. You MUST output exclusively valid JSON without markdown wrapping.
4. Output Schema: {"risk_band": "LOW"|"MEDIUM"|"HIGH", "flags": [strings], "recommendation": "approve"|"verify"|"reject", "confidence": float}
`,
		s.Name, s.DeclaredIncome, s.LoanPurpose, s.Employment, s.VerbalConsent,
	)
}

func mockExecuteLLMRequest(sessionID string, prompt string) *LLMEvaluationResult {
	log.Printf("[LLM-%s] Generating structural response. Payload length: %d bytes. Simulating Gemini latency...", sessionID, len(prompt))
	time.Sleep(1200 * time.Millisecond)

	// User specifically requested this rigid output mock array
	rawJSONMock := `{
		"risk_band": "MEDIUM",
		"flags": ["age_mismatch", "location_mismatch"],
		"recommendation": "verify",
		"confidence": 0.74
	}`

	var result LLMEvaluationResult
	err := json.Unmarshal([]byte(rawJSONMock), &result)
	if err != nil {
		log.Printf("[LLM-%s] Warning: Simulated LLM Parsing Error: %v", sessionID, err)
	}

	log.Printf("[LLM-%s] Evaluated completed. Evaluated Final Risk Tier: %s", sessionID, result.RiskBand)
	return &result
}
