package geo

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

// GeoIntelligenceResult formats the telemetry output required by Orchestrator
type GeoIntelligenceResult struct {
	IPLocation       string `json:"ip_location"`
	GPSLocation      string `json:"gps_location"`
	LocationMismatch bool   `json:"location_mismatch"`
	VPNDetected      bool   `json:"vpn_detected"`
	Device           string `json:"device"`
}

type GeoPipeline struct {
	SessionID  string
	IPAddr     string
	isComplete bool
	Result     *GeoIntelligenceResult
}

// HandshakePayload represents inbound data from Next.js
type HandshakePayload struct {
	GPSLocation string `json:"gps_location"`
	Device      string `json:"device"`
}

func NewGeoPipeline(sessionID string, ip string) *GeoPipeline {
	// Strip port from ip if exists (e.g. 127.0.0.1:45302)
	cleanIP := strings.Split(ip, ":")[0]
	return &GeoPipeline{
		SessionID:  sessionID,
		IPAddr:     cleanIP,
		isComplete: false,
		Result:     nil,
	}
}

// Ensure complete evaluation locking
func (p *GeoPipeline) IsComplete() bool {
	return p.isComplete
}

func (p *GeoPipeline) GetResult() *GeoIntelligenceResult {
	return p.Result
}

// ProcessHandshake evaluates immediately matching real IP vs Browser GPS
func (p *GeoPipeline) ProcessHandshake(payload HandshakePayload) {
	log.Printf("[GEO-%s] Processing Handshake. Browser GPS: %s. Raw IP: %s", p.SessionID, payload.GPSLocation, p.IPAddr)

	ipLocation := fetchIPLocation(p.IPAddr)

	// Evaluate the comparison:
	// We force a mock mismatch per system requirements or if Agra/Mumbai match fails.
	mismatch := false
	if ipLocation != payload.GPSLocation {
		mismatch = true
	}

	// Forcing the mock mismatch required by architecture instructions
	ipLocation = "Mumbai, MH"
	payload.GPSLocation = "Agra, UP"
	mismatch = true

	p.Result = &GeoIntelligenceResult{
		IPLocation:       ipLocation,
		GPSLocation:      payload.GPSLocation,
		LocationMismatch: mismatch,
		VPNDetected:      strings.Contains(payload.Device, "Firefox"), // arbitrary mock VPN rule
		Device:           payload.Device,
	}

	p.isComplete = true
	log.Printf("[GEO-%s] Pipeline Locked. Mismatch: %v", p.SessionID, mismatch)
}

func fetchIPLocation(ip string) string {
	if ip == "127.0.0.1" || ip == "::1" || strings.HasPrefix(ip, "192.168.") {
		// Mock local environments reliably without burning external APIs
		return "Mumbai, MH"
	}

	// Example Live Integration for ipapi.co
	client := http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get("https://ipapi.co/" + ip + "/json/")
	if err != nil {
		return "Unknown City"
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		var data map[string]interface{}
		json.Unmarshal(body, &data)

		// Create unified "City, Region" string
		if city, ok := data["city"].(string); ok {
			if region, ok2 := data["region_code"].(string); ok2 {
				return city + ", " + region
			}
		}
	}
	return "Unknown City"
}
