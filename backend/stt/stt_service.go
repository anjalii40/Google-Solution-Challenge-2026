package stt

import (
	"log"
	"time"
)

type ExtractedData struct {
	Name           string `json:"name"`
	DeclaredIncome int    `json:"declared_income"`
	LoanPurpose    string `json:"loan_purpose"`
	Employment     string `json:"employment"`
	VerbalConsent  bool   `json:"verbal_consent"`
}

type STTPipeline struct {
	SessionID  string
	AudioCount int
	isComplete bool
	Result     *ExtractedData
}

func NewSTTPipeline(sessionID string) *STTPipeline {
	return &STTPipeline{
		SessionID:  sessionID,
		AudioCount: 0,
		isComplete: false,
		Result:     nil,
	}
}

func (p *STTPipeline) ProcessChunk(chunk []byte) {
	if p.isComplete {
		return
	}

	p.AudioCount++
	log.Printf("[STT-%s] Streaming chunk %d to Deepgram API (simulated)...", p.SessionID, p.AudioCount)

	if p.AudioCount >= 5 {
		p.finalizeExtraction()
	}
}

func (p *STTPipeline) IsComplete() bool {
	return p.isComplete
}

func (p *STTPipeline) GetResult() *ExtractedData {
	return p.Result
}

func (p *STTPipeline) finalizeExtraction() {
	log.Printf("[STT-%s] Closing remote stream. Triggering Deepgram parser/extraction...", p.SessionID)
	time.Sleep(500 * time.Millisecond)

	// **FORCING `VerbalConsent: false` TO EXPLICITLY TEST MODULE 7 HARD REJECTION UI**
	p.Result = &ExtractedData{
		Name:           "Rahul Sharma",
		DeclaredIncome: 45000,
		LoanPurpose:    "business expansion",
		Employment:     "self-employed",
		VerbalConsent:  false, // <-- SET TO FALSE FOR TESTING
	}
	
	p.isComplete = true
	log.Printf("[STT-%s] Extraction Complete", p.SessionID)
}
