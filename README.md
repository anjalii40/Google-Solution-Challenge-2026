# LoanCall AI

**AI Video Call-Based Loan Onboarding in Real Time**

---

## 1. Project Title + One-line pitch

**LoanCall AI** — Agentic AI system for instant, compliant loan onboarding via live video call. No paperwork. No manual KYC.

---

## 2. Problem Statement

- Traditional loan onboarding is slow, manual, and error-prone
- KYC (Know Your Customer) requires physical documents and human verification
- Fraud risk is high with static checks
- Users expect instant, digital-first experiences

---

## 3. Solution Overview

- Real-time video call replaces KYC forms and manual checks
- Parallel AI pipelines extract, verify, and structure user data during the call
- Instant risk scoring and loan offer generation
- Designed for hackathons and production-grade extensibility

---

## 4. System Architecture

**Flow:**
1. User joins a secure video call (WebRTC)
2. Audio/video streams sent to backend (Go WebSocket server)
3. AI pipelines run in parallel:
		- Speech-to-Text (Deepgram): Extracts user details, consent
		- Computer Vision (Google Vision): Face match, age estimation
		- Geo/Device Intelligence: IP vs GPS, VPN detection
		- LLM (Gemini): Structures and normalizes all data
		- Risk Engine: Calculates score
		- Offer Generator: Produces eligibility
4. Results streamed back to frontend in real time

```
User <-> Next.js/WebRTC <-> Go WebSocket Server <-> AI Pipelines <-> PostgreSQL/Cloud Storage
```

---

## 5. Key Features

- Real-time KYC via video call
- Parallel AI pipelines (STT, CV, Geo, LLM)
- Fraud detection (geo/device mismatch, VPN)
- Deterministic risk scoring
- Instant loan offer generation
- Cloud-native deployment (Docker, GCP)

---

## 6. Tech Stack

- **Frontend:** Next.js, WebRTC
- **Backend:** Go (WebSocket), Node.js
- **AI:** Gemini API, Deepgram, Google Vision
- **Database:** PostgreSQL
- **Deployment:** Docker, Google Cloud Run
- **Storage:** Google Cloud Storage

---

## 7. How It Works (Step-by-Step)

1. User starts onboarding, joins video call
2. Audio/video streamed to backend
3. STT extracts name, address, consent
4. CV verifies face, estimates age
5. Geo/Device checks for fraud (IP ≠ GPS, VPN)
6. LLM normalizes all extracted data
7. Risk Engine scores user
8. Offer Generator returns eligibility
9. Results shown instantly to user

---

## 8. Risk Engine Logic

- Deterministic, rule-based scoring
- Example rules:
		- Age < 21: Reject
		- Geo mismatch or VPN: High risk
		- Missing consent: Reject
		- All checks pass: Score = 100
- Example output:

```json
{
	"age": 24,
	"geo_match": true,
	"vpn": false,
	"consent": true,
	"score": 100,
	"eligible": true
}
```

---

## 9. Installation & Setup

```sh
# Clone repo
$ git clone https://github.com/anjalii40/Google-Solution-Challenge-2026.git
$ cd Google-Solution-Challenge-2026

# Build and run all services with Docker Compose
$ docker-compose up --build
```

- Frontend: http://localhost:3000
- Backend: ws://localhost:8080

---

## 10. API / WebSocket Flow

- Frontend connects to backend via WebSocket
- Streams audio/video frames
- Receives structured JSON updates in real time

**Sample WebSocket message:**

```json
{
	"event": "risk_update",
	"data": {
		"score": 85,
		"eligible": true,
		"offer": "INR 50,000 @ 14%"
	}
}
```

---

## 11. Demo Flow (60–90 seconds)

1. User clicks "Start Loan Onboarding"
2. Joins live video call
3. Speaks name, address, gives consent
4. System verifies face, checks geo/device
5. Risk Engine scores, Offer Generator responds
6. User sees eligibility and offer instantly

---

## 12. Future Improvements

- Add OCR for ID document extraction
- Support for multi-language STT
- Advanced fraud analytics (device fingerprinting)
- UI/UX polish for production
- Integration with credit bureaus

---

## 13. Team

- Anjali Prajapati (Lead Engineer)
- [Add more team members here]

## Getting Started

First, run the development server:

```bash
npm run dev
# or
yarn dev
# orP
pnpm dev
# or
bun dev
```

Open [http://localhost:3000](http://localhost:3000) with your browser to see the result.

You can start editing the page by modifying `app/page.tsx`. The page auto-updates as you edit the file.

This project uses [`next/font`](https://nextjs.org/docs/app/building-your-application/optimizing/fonts) to automatically optimize and load [Geist](https://vercel.com/font), a new font family for Vercel.

## Learn More

To learn more about Next.js, take a look at the following resources:

- [Next.js Documentation](https://nextjs.org/docs) - learn about Next.js features and API.
- [Learn Next.js](https://nextjs.org/learn) - an interactive Next.js tutorial.

You can check out [the Next.js GitHub repository](https://github.com/vercel/next.js) - your feedback and contributions are welcome!

## Deploy on Vercel

The easiest way to deploy your Next.js app is to use the [Vercel Platform](https://vercel.com/new?utm_medium=default-template&filter=next.js&utm_source=create-next-app&utm_campaign=create-next-app-readme) from the creators of Next.js.

Check out our [Next.js deployment documentation](https://nextjs.org/docs/app/building-your-application/deploying) for more details.
