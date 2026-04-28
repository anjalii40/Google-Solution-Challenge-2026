package risk

import (
	"log"

	"backend/cv"
	"backend/geo"
	"backend/llm"
	"backend/stt"
)

type RiskEvaluation struct {
	FinalScore      int
	RiskTier        string
	Flags           []string
	IsRejected      bool
	RejectionReason string
}

// CalculateEngine merges inputs procedurally into a deterministic score bounding tier.
func CalculateEngine(sttData *stt.ExtractedData, cvData *cv.AgeEstimationResult, geoData *geo.GeoIntelligenceResult, llmData *llm.LLMEvaluationResult) *RiskEvaluation {
	log.Println("[RiskEngine] Initiating deterministic evaluation matrix...")

	evaluation := &RiskEvaluation{
		FinalScore: 100, // Base
		Flags:      llmData.Flags, // Absorb LLM insights immediately
		IsRejected: false,
	}

	// 1. FATAL BINARY GATE: VERBAL CONSENT
	if !sttData.VerbalConsent {
		evaluation.IsRejected = true
		evaluation.RejectionReason = "No verbal consent provided (Legal Requirement Failure)"
		log.Println("[RiskEngine] FATAL FLAG: Missing Verbal Consent. Hard rejecting application.")
		return evaluation
	}

	// 2. PENALTY CALCULATIONS (Deterministic Rules)
	if cvData.Flag {
		log.Println("[RiskEngine] Applying Penalty: Age Mismatch (-30)")
		evaluation.FinalScore -= 30
		appendUniqueFlag(evaluation, "age_mismatch")
	}

	if geoData.LocationMismatch {
		log.Println("[RiskEngine] Applying Penalty: Location Mismatch (-20)")
		evaluation.FinalScore -= 20
		appendUniqueFlag(evaluation, "location_mismatch")
	}

	if geoData.VPNDetected {
		log.Println("[RiskEngine] Applying Penalty: VPN Detected (-25)")
		evaluation.FinalScore -= 25
		appendUniqueFlag(evaluation, "vpn_detected")
	}

	if sttData.DeclaredIncome < 30000 {
		log.Println("[RiskEngine] Applying Penalty: Income Threshold Failure (-15)")
		evaluation.FinalScore -= 15
		appendUniqueFlag(evaluation, "low_income_threshold")
	}

	// 3. TIER MAPPING
	if evaluation.FinalScore >= 80 {
		evaluation.RiskTier = "LOW"
	} else if evaluation.FinalScore >= 60 {
		evaluation.RiskTier = "MEDIUM"
	} else {
		evaluation.RiskTier = "HIGH"
	}

	log.Printf("[RiskEngine] Risk Evaluation Complete. Final Score: %d. Tier Maps: %s", evaluation.FinalScore, evaluation.RiskTier)

	return evaluation
}

// Utility mapper to keep flags list clean and deduplicated
func appendUniqueFlag(e *RiskEvaluation, flag string) {
	for _, f := range e.Flags {
		if f == flag {
			return
		}
	}
	e.Flags = append(e.Flags, flag)
}
