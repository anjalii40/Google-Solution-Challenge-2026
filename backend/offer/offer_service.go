package offer

import (
	"log"
	"math"
)

// GeneratedOffer isolates final parameters
type GeneratedOffer struct {
	Amount       int
	TenureMonths int
	EMI          int
	Rate         float64
}

// CalculateOffer evaluates deterministic boundaries mapping to mathematical rules.
// LOW risk -> up to 5x monthly income, 48 months
// MEDIUM risk -> up to 3x monthly income, 24 months
func CalculateOffer(riskTier string, monthlyIncome int) *GeneratedOffer {
	log.Printf("[OfferEngine] Processing dynamic evaluation for Tier %s. Declared Income Base: %d", riskTier, monthlyIncome)

	if riskTier == "HIGH" || riskTier == "REJECTED" {
		return nil // Highly bounded UI systems prevent passing this, but backend rules exist!
	}

	offer := &GeneratedOffer{}

	if riskTier == "LOW" {
		offer.Amount = monthlyIncome * 5
		offer.TenureMonths = 48
		offer.Rate = 10.5
	} else if riskTier == "MEDIUM" {
		offer.Amount = monthlyIncome * 3
		offer.TenureMonths = 24
		offer.Rate = 14.5
	}

	// Very simple EMI proxy calculation: Principal / Tenure + Flat rate simple proxy.
	// True systems execute complex amortizations, this demonstrates sufficient numerical mapping.
	monthlyPrincipal := float64(offer.Amount) / float64(offer.TenureMonths)
	interestChunk := (float64(offer.Amount) * (offer.Rate / 100)) / float64(offer.TenureMonths)

	offer.EMI = int(math.Round(monthlyPrincipal + interestChunk))

	log.Printf("[OfferEngine] Bound generation successful. Amount: %d, EMI: %d", offer.Amount, offer.EMI)
	return offer
}
