package cv

import (
	"log"
	"time"
)

// AgeEstimationResult tracks the demographic analysis returned.
type AgeEstimationResult struct {
	EstimatedAgeRange []int  `json:"estimated_age_range"`
	DeclaredAge       int    `json:"declared_age"`
	Flag              bool   `json:"flag"`
	FlagReason        string `json:"flag_reason"` // e.g. "age_mismatch"
}

// CVPipeline struct models the connection and incoming visual buffer.
type CVPipeline struct {
	SessionID  string
	FrameCount int
	isComplete bool
	Result     *AgeEstimationResult
}

func NewCVPipeline(sessionID string) *CVPipeline {
	return &CVPipeline{
		SessionID:  sessionID,
		FrameCount: 0,
		isComplete: false,
		Result:     nil,
	}
}

// ProcessFrame ingests base64 packets and evaluates the image.
func (p *CVPipeline) ProcessFrame(frame []byte) {
	if p.isComplete {
		return
	}

	p.FrameCount++
	log.Printf("[CV-%s] Streaming video frame %d for evaluation...", p.SessionID, p.FrameCount)

	// Simulate CV locking its result at frame 5
	if p.FrameCount >= 5 {
		p.finalizeExtraction()
	}
}

// IsComplete ensures orchestrators know CV processing is locked.
func (p *CVPipeline) IsComplete() bool {
	return p.isComplete
}

// GetResult retrieves the analyzed JSON payload containing flags.
func (p *CVPipeline) GetResult() *AgeEstimationResult {
	return p.Result
}

func (p *CVPipeline) finalizeExtraction() {
	log.Printf("[CV-%s] Triggering Google Vision + DeepFace evaluations...", p.SessionID)
	time.Sleep(600 * time.Millisecond)

	// Mocking the EXACT failure condition mapped in design requirements:
	// They declared 42, but DeepFace estimates 28-35. (Difference > 10yrs)
	p.Result = &AgeEstimationResult{
		EstimatedAgeRange: []int{28, 35},
		DeclaredAge:       42,
		Flag:              true,
		FlagReason:        "age_mismatch",
	}

	p.isComplete = true
	log.Printf("[CV-%s] Age Estimation Complete. FLAG = %v", p.SessionID, p.Result.Flag)
}
