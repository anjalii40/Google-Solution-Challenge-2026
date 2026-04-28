package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"backend/cv"
	"backend/geo"
	"backend/llm"
	"backend/offer"
	"backend/risk"
	"backend/stt"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	_ "github.com/lib/pq"
)

var db *sql.DB

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for dev
	},
}

var activeSessions = sync.Map{}

type Session struct {
	ID   string
	Conn *websocket.Conn
	STT  *stt.STTPipeline
	CV   *cv.CVPipeline
	GEO  *geo.GeoPipeline
}

type WSMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

type OfferPayload struct {
	Status          string   `json:"status"` // "APPROVED" | "MANUAL_REVIEW" | "REJECTED"
	Reason          string   `json:"reason,omitempty"`
	Amount          int      `json:"amount,omitempty"`
	EMI             int      `json:"emi,omitempty"`
	Tenure          int      `json:"tenure,omitempty"`
	InterestRate    float64  `json:"interestRate,omitempty"`
	RiskTier        string   `json:"risk_tier,omitempty"`
	Flags           []string `json:"flags,omitempty"`
	ManualReviewReq bool     `json:"manual_review_required,omitempty"`
}

func main() {
	initDB()
	http.HandleFunc("/ws/onboard", handleWebSocket)

	port := "8080"
	log.Printf("Server starting on :%s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func initDB() {
	host := os.Getenv("DB_HOST")
	if host == "" {
		host = "localhost"
	}
	port := os.Getenv("DB_PORT")
	if port == "" {
		port = "5432"
	}
	user := os.Getenv("DB_USER")
	if user == "" {
		user = "admin"
	}
	password := os.Getenv("DB_PASSWORD")
	if password == "" {
		password = "password"
	}
	dbname := os.Getenv("DB_NAME")
	if dbname == "" {
		dbname = "onboard_db"
	}

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)

	var err error
	for i := 0; i < 5; i++ {
		db, err = sql.Open("postgres", connStr)
		if err == nil {
			err = db.Ping()
			if err == nil {
				log.Println("Connected to PostgreSQL successfully")
				return
			}
		}
		log.Println("Waiting for PostgreSQL to be ready...")
		time.Sleep(2 * time.Second)
	}
	log.Fatalf("Failed to connect to database: %v", err)
}

func logToDB(sessionID string, message string) {
	if db != nil {
		_, err := db.Exec("INSERT INTO session_logs (session_id, log_message) VALUES ($1, $2)", sessionID, message)
		if err != nil {
			log.Printf("DB Log Err: %v", err)
		}
	}
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}
	defer conn.Close()

	sessionID := uuid.New().String()
	
	session := &Session{
		ID:   sessionID,
		Conn: conn,
		STT:  stt.NewSTTPipeline(sessionID),
		CV:   cv.NewCVPipeline(sessionID),
		GEO:  geo.NewGeoPipeline(sessionID, r.RemoteAddr),
	}
	activeSessions.Store(sessionID, session)

	_, err = db.Exec("INSERT INTO sessions (session_id, status) VALUES ($1, $2)", sessionID, "STARTED")
	if err != nil {
		log.Println("Failed to create session in DB:", err)
	}
	log.Printf("New Session %s started. Raw IP Bound: %s\n", sessionID, r.RemoteAddr)
	logToDB(sessionID, "Session initialized")

	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Session %s disconnected: %v\n", sessionID, err)
			activeSessions.Delete(sessionID)
			db.Exec("UPDATE sessions SET status = 'CLOSED', updated_at = CURRENT_TIMESTAMP WHERE session_id = $1", sessionID)
			return
		}

		if messageType == websocket.TextMessage {
			var msg WSMessage
			if err := json.Unmarshal(p, &msg); err != nil {
				continue
			}

			if msg.Type == "device_handshake" {
				dataMap, ok := msg.Data.(map[string]interface{})
				if !ok {
					continue
				}
				
				payload := geo.HandshakePayload{
					GPSLocation: dataMap["gps_location"].(string),
					Device:      dataMap["device"].(string),
				}
				session.GEO.ProcessHandshake(payload)
				logToDB(sessionID, "Processed Geo Device Handshake")
				
				checkOrchestrator(session)
				
			} else if msg.Type == "audio_chunk" {
				session.STT.ProcessChunk(p)
				checkOrchestrator(session)
			} else if msg.Type == "video_frame" {
				session.CV.ProcessFrame(p)
				checkOrchestrator(session)
			}
		}
	}
}

func checkOrchestrator(s *Session) {
	if s.CV.FrameCount == 1 && s.STT.AudioCount == 0 {
		sendStatus(s.Conn, "Verifying identity paths...")
	}

	if s.GEO.IsComplete() && s.CV.IsComplete() && s.STT.IsComplete() {
		if _, ok := activeSessions.Load(s.ID); !ok {
			return
		}
		activeSessions.Delete(s.ID) 
		
		log.Printf("[Orchestrator] All pipelines complete for %s", s.ID)
		sendStatus(s.Conn, "Synthesizing multimodal models for LLM Context Prompt...")
		
		extractedSTT := s.STT.GetResult()
		extractedCV  := s.CV.GetResult()
		extractedGEO := s.GEO.GetResult()

		// 1. LLM SYNTHESIS
		llmResult := llm.EvaluateRisk(s.ID, extractedSTT)

		// 2. DETERMINISTIC RISK RULES OVERRIDE
		sendStatus(s.Conn, "Executing Deterministic Policy Rule Engine...")
		finalEvaluation := risk.CalculateEngine(extractedSTT, extractedCV, extractedGEO, llmResult)

		// Logs
		logToDB(s.ID, fmt.Sprintf("Calculated Engine Base Score: %d | Final Tier: %s", finalEvaluation.FinalScore, finalEvaluation.RiskTier))
		for _, f := range finalEvaluation.Flags {
			logToDB(s.ID, fmt.Sprintf("Assigned Structural Flag: %s", f))
		}
		
		go func() {
			time.Sleep(1 * time.Second)
			
			var offerPayload OfferPayload
			dbStatusRecord := ""

			if finalEvaluation.IsRejected {
				// Condition 1: RED HARD REJECT
				offerPayload = OfferPayload{
					Status: "REJECTED",
					Reason: finalEvaluation.RejectionReason,
				}
				dbStatusRecord = "REJECTED"
			} else if finalEvaluation.RiskTier == "HIGH" || llmResult.Recommendation == "verify" || len(finalEvaluation.Flags) > 0 {
				// Condition 2: YELLOW MANUAL REVIEW
				offerPayload = OfferPayload{
					Status:          "MANUAL_REVIEW",
					RiskTier:        finalEvaluation.RiskTier,
					Flags:           finalEvaluation.Flags,
					ManualReviewReq: true,
				}
				dbStatusRecord = "MANUAL_REVIEW_FLAGGED"
			} else {
				// Condition 3: GREEN OFFER GENERATED
				generated := offer.CalculateOffer(finalEvaluation.RiskTier, extractedSTT.DeclaredIncome)
				if generated == nil {
					// Fallback for unexpected missing generation rules
					generated = &offer.GeneratedOffer{Amount: 500000, EMI: 12500, TenureMonths: 48, Rate: 10.5}
				}
				offerPayload = OfferPayload{
					Status:       "APPROVED",
					Amount:       generated.Amount,
					EMI:          generated.EMI,
					Tenure:       generated.TenureMonths,
					InterestRate: generated.Rate,
					RiskTier:     finalEvaluation.RiskTier,
				}
				dbStatusRecord = "OFFER_DELIVERED"
			}

			offerJSON, _ := json.Marshal(offerPayload.Flags)
			if offerPayload.Flags == nil {
				offerJSON = []byte("[]")
			}
			
			// Module 8: Audit Log insertion
			go func() {
				// STT
				db.Exec(`INSERT INTO transcripts (session_id, name, declared_income, loan_purpose, employment, verbal_consent) 
						 VALUES ($1, $2, $3, $4, $5, $6)`,
					s.ID, extractedSTT.Name, extractedSTT.DeclaredIncome, extractedSTT.LoanPurpose, extractedSTT.Employment, extractedSTT.VerbalConsent)
				
				// CV
				estMin, estMax := 0, 0
				if len(extractedCV.EstimatedAgeRange) >= 2 {
					estMin = extractedCV.EstimatedAgeRange[0]
					estMax = extractedCV.EstimatedAgeRange[1]
				}
				db.Exec(`INSERT INTO cv_results (session_id, estimated_age_min, estimated_age_max, declared_age, flag, flag_reason) 
						 VALUES ($1, $2, $3, $4, $5, $6)`,
					s.ID, estMin, estMax, extractedCV.DeclaredAge, extractedCV.Flag, extractedCV.FlagReason)
				
				// GEO
				db.Exec(`INSERT INTO geo_results (session_id, ip_location, gps_location, location_mismatch, vpn_detected, device) 
						 VALUES ($1, $2, $3, $4, $5, $6)`,
					s.ID, extractedGEO.IPLocation, extractedGEO.GPSLocation, extractedGEO.LocationMismatch, extractedGEO.VPNDetected, extractedGEO.Device)

				// LLM
				llmFlagsJSON, _ := json.Marshal(llmResult.Flags)
				if llmResult.Flags == nil {
					llmFlagsJSON = []byte("[]")
				}
				db.Exec(`INSERT INTO llm_outputs (session_id, risk_band, flags, recommendation, confidence) 
						 VALUES ($1, $2, $3, $4, $5)`,
					s.ID, llmResult.RiskBand, llmFlagsJSON, llmResult.Recommendation, llmResult.Confidence)
				
				// Offer
				db.Exec(`INSERT INTO loan_offers (session_id, status, reason, amount, emi, tenure, interest_rate, risk_tier, flags, manual_review_required) 
						 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
					s.ID, offerPayload.Status, offerPayload.Reason, offerPayload.Amount, offerPayload.EMI, offerPayload.Tenure, offerPayload.InterestRate, offerPayload.RiskTier, offerJSON, offerPayload.ManualReviewReq)
			}()

			sendOffer(s.Conn, offerPayload)
			db.Exec("UPDATE sessions SET status = $2, updated_at = CURRENT_TIMESTAMP WHERE session_id = $1", s.ID, dbStatusRecord)
			logToDB(s.ID, "Engine dispatched evaluated constraint transaction")
		}()
	}
}

func sendStatus(conn *websocket.Conn, msg string) {
	payload := WSMessage{
		Type: "status_update",
		Data: map[string]string{"message": msg},
	}
	conn.WriteJSON(payload)
}

func sendOffer(conn *websocket.Conn, offer OfferPayload) {
	payload := WSMessage{
		Type: "offer_received",
		Data: offer,
	}
	conn.WriteJSON(payload)
}
