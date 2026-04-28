'use client';

import { useState, useRef, useEffect } from 'react';
import { Camera, Mic, ShieldCheck, CheckCircle, Video, CreditCard, AlertTriangle, UserX, XCircle } from 'lucide-react';
import { motion, AnimatePresence } from 'framer-motion';

// Types
type OfferPayload = {
  status: string; // "APPROVED" | "MANUAL_REVIEW" | "REJECTED"
  reason?: string;
  amount?: number;
  emi?: number;
  tenure?: number;
  interestRate?: number;
  risk_tier?: string;
  flags?: string[];
  manual_review_required?: boolean;
}

export default function Home() {
  const [sessionState, setSessionState] = useState<'IDLE' | 'PERMISSIONS' | 'ACTIVE' | 'OFFER'>('IDLE');
  const [statusMessage, setStatusMessage] = useState('Initializing secure session...');
  const [offer, setOffer] = useState<OfferPayload | null>(null);
  
  const videoRef = useRef<HTMLVideoElement>(null);
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const mediaRecorderRef = useRef<MediaRecorder | null>(null);
  const wsRef = useRef<WebSocket | null>(null);

  const startSession = async () => {
    setSessionState('PERMISSIONS');
    
    try {
      const ws = new WebSocket('ws://localhost:8080/ws/onboard');
      wsRef.current = ws;
      
      ws.onopen = () => {
        console.log('Connected to Go Backend WS via Native WebSocket');
        
        const ua = navigator.userAgent;
        let mockFallbackGPS = "Agra, UP";
        
        if ("geolocation" in navigator) {
          navigator.geolocation.getCurrentPosition(
             (pos) => { ws.send(JSON.stringify({ type: 'device_handshake', data: { gps_location: mockFallbackGPS, device: ua } })); },
             (err) => { ws.send(JSON.stringify({ type: 'device_handshake', data: { gps_location: mockFallbackGPS, device: ua } })); }
          );
        } else {
           ws.send(JSON.stringify({ type: 'device_handshake', data: { gps_location: mockFallbackGPS, device: ua } }));
        }
      };

      ws.onmessage = (event) => {
        try {
          const payload = JSON.parse(event.data);
          if (payload.type === 'status_update') {
            setStatusMessage(payload.data.message);
          } else if (payload.type === 'offer_received') {
            console.log("Offer Evaluation received!", payload.data);
            setOffer(payload.data);
            setSessionState('OFFER');
            stopMediaStreams();
          }
        } catch(e) {
          console.error("Failed to parse WS message", event.data);
        }
      };

      const stream = await navigator.mediaDevices.getUserMedia({ video: { facingMode: 'user', width: { ideal: 1280 }, height: { ideal: 720 } }, audio: true });
      if (videoRef.current) { videoRef.current.srcObject = stream; }
      setSessionState('ACTIVE');

      const mimeType = MediaRecorder.isTypeSupported('audio/webm') ? 'audio/webm' : 'audio/mp4';
      const recorder = new MediaRecorder(stream, { mimeType });
      mediaRecorderRef.current = recorder;

      recorder.ondataavailable = async (event) => {
        if (event.data.size > 0 && wsRef.current?.readyState === WebSocket.OPEN) {
          const reader = new FileReader();
          reader.onloadend = () => { wsRef.current?.send(JSON.stringify({ type: 'audio_chunk', data: reader.result })); };
          reader.readAsDataURL(event.data);
        }
      };
      recorder.start(3000);

    } catch (err) {
      console.error('Error accessing media', err);
      alert('Camera & Microphone access is required.');
      setSessionState('IDLE');
    }
  };

  useEffect(() => {
    let frameInterval: NodeJS.Timeout;
    if (sessionState === 'ACTIVE') {
      frameInterval = setInterval(() => {
        if (videoRef.current && canvasRef.current && wsRef.current?.readyState === WebSocket.OPEN) {
           const video = videoRef.current;
           const canvas = canvasRef.current;
           const ctx = canvas.getContext('2d');
           if (ctx) {
             canvas.width = video.videoWidth;
             canvas.height = video.videoHeight;
             ctx.drawImage(video, 0, 0, canvas.width, canvas.height);
             wsRef.current.send(JSON.stringify({ type: 'video_frame', data: canvas.toDataURL('image/jpeg', 0.8) }));
           }
        }
      }, 5000);
    }
    return () => { if (frameInterval) clearInterval(frameInterval); };
  }, [sessionState]);

  const stopMediaStreams = () => {
    if (mediaRecorderRef.current) { mediaRecorderRef.current.stop(); }
    if (videoRef.current?.srcObject) {
      const tracks = (videoRef.current.srcObject as MediaStream).getTracks();
      tracks.forEach(track => track.stop());
    }
    if (wsRef.current) { wsRef.current.close(); }
  };

  const formatFlags = (flag: string) => {
    switch (flag) {
      case "age_mismatch": return "Age profile mismatch detected";
      case "location_mismatch": return "GPS validation sequence mismatch";
      case "vpn_detected": return "Encrypted network proxy usage observed";
      default: return `Automatic flag raised (${flag})`;
    }
  }

  // Helper determining border coloration natively
  const evaluateBorder = () => {
    if (offer?.status === "REJECTED") return { borderTop: '6px solid #dc2626' } // Red
    if (offer?.status === "MANUAL_REVIEW") return { borderTop: '6px solid #f59e0b' } // Amber/Yellow
    return { borderTop: '6px solid #10b981' } // Green
  }

  return (
    <main className="container">
      <canvas ref={canvasRef} style={{ display: 'none' }} />
      <AnimatePresence mode="wait">
        
        {sessionState === 'IDLE' && (
           <motion.div key="idle" initial={{ opacity: 0, y: 20 }} animate={{ opacity: 1, y: 0 }} exit={{ opacity: 0, y: -20 }} className="glass-card">
            <div style={{ background: 'rgba(59, 130, 246, 0.2)', padding: '1rem', borderRadius: '50%' }}>
              <ShieldCheck size={48} color="#3b82f6" />
            </div>
            <h1 className="title" style={{ fontSize: '2rem', marginBottom: '0.5rem' }}>Agentic Onboarding</h1>
            <p className="subtitle" style={{ marginBottom: '2rem' }}>We need to verify your identity to generate a personalized loan offer.</p>
            <button className="btn-primary" onClick={startSession}><Video size={20} />Start Secure Session</button>
          </motion.div>
        )}

        {sessionState === 'PERMISSIONS' && (
           <motion.div key="permissions" initial={{ opacity: 0, scale: 0.95 }} animate={{ opacity: 1, scale: 1 }} exit={{ opacity: 0, scale: 1.05 }} className="glass-card">
            <Camera className="pulse" size={40} color="#94a3b8" />
            <h2 style={{ fontSize: '1.4rem' }}>Requesting Access...</h2>
            <div style={{ display: 'flex', gap: '1rem' }}><Camera size={24} /> <Mic size={24} /></div>
          </motion.div>
        )}

        {sessionState === 'ACTIVE' && (
          <motion.div key="active" initial={{ opacity: 0, scale: 0.9 }} animate={{ opacity: 1, scale: 1 }} exit={{ opacity: 0, scale: 1.1 }} className="video-container">
            <video ref={videoRef} autoPlay playsInline muted className="live-video" />
            <div className="status-overlay">
              <div className="status-badge"><div className="pulse"></div>LIVE SECURE STREAM</div>
              <div className="message-box fade-in" key={statusMessage}>{statusMessage}</div>
            </div>
          </motion.div>
        )}

        {sessionState === 'OFFER' && offer && (
          <motion.div key="offer" initial={{ opacity: 0, y: 50 }} animate={{ opacity: 1, y: 0 }} transition={{ type: 'spring' }} className="offer-card" style={evaluateBorder()}>
            
            {/* STATE 1: HARD REJECTED (🔴) */}
            {offer.status === "REJECTED" && (
              <>
                 <XCircle size={64} color="#dc2626" />
                 <h2 style={{ fontSize: '2rem', marginTop: '1rem', color: '#f8fafc', fontWeight: 800 }}>Application Rejected</h2>
                 <p style={{ color: 'var(--text-secondary)', textAlign: 'center', marginBottom: '1.5rem', fontWeight: 500 }}>
                   Unmet compliance requirements.
                 </p>
                 
                 <div style={{ background: 'rgba(220, 38, 38, 0.1)', border: '1px solid rgba(220, 38, 38, 0.4)', padding: '1.5rem', borderRadius: '12px', width: '100%', marginBottom: '2rem' }}>
                   <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', color: '#fca5a5', fontWeight: 600, marginBottom: '0.5rem' }}>
                     Declination Reason:
                   </div>
                   <div style={{ color: '#white', fontSize: '1.1rem', fontWeight: 500 }}>
                     {offer.reason}
                   </div>
                 </div>
                 
                 <button className="btn-primary" style={{ width: '100%', justifyContent: 'center', background: '#334155' }} onClick={() => window.location.reload()}>
                   Return to Safety Hub
                 </button>
              </>
            )}

            {/* STATE 2: MANUAL REVIEW (🟡) */}
            {offer.status === "MANUAL_REVIEW" && (
              <>
                 <UserX size={56} color="#f59e0b" />
                 <h2 style={{ fontSize: '1.8rem', marginTop: '1rem', color: '#f8fafc' }}>Review Required</h2>
                 <p style={{ color: 'var(--text-secondary)', textAlign: 'center', marginBottom: '1.5rem' }}>
                   Our system detected inconsistencies during verification. A human agent will review your application shortly.
                 </p>
                 
                 <div style={{ background: 'rgba(245, 158, 11, 0.1)', border: '1px solid rgba(245, 158, 11, 0.3)', padding: '1rem', borderRadius: '12px', width: '100%', marginBottom: '2rem' }}>
                   <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', color: '#fcd34d', fontWeight: 600, marginBottom: '0.5rem' }}>
                     <AlertTriangle size={18} /> Engine Flags:
                   </div>
                   <ul style={{ color: 'var(--text-secondary)', marginLeft: '1.5rem', fontSize: '0.9rem' }}>
                     {offer.flags?.map((flag, idx) => (
                       <li key={idx}>{formatFlags(flag)}</li>
                     ))}
                   </ul>
                 </div>
                 
                 <button className="btn-primary" style={{ width: '100%', justifyContent: 'center', background: '#334155' }} onClick={() => window.location.reload()}>
                   Return Home
                 </button>
              </>
            )}

            {/* STATE 3: APPROVED (🟢) */}
            {offer.status === "APPROVED" && (
              <>
                <CheckCircle size={56} color="#10b981" />
                <h2 style={{ fontSize: '1.8rem', marginTop: '1rem' }}>Identity Verified!</h2>
                <p style={{ color: 'var(--text-secondary)' }}>Here is your tailored instant loan offer.</p>
                
                <div className="offer-amount">
                  ₹{offer?.amount?.toLocaleString('en-IN')}
                </div>
                
                <div className="offer-details">
                  <div className="detail-item">
                    <span className="detail-label">EMI</span>
                    <span className="detail-value">₹{offer?.emi?.toLocaleString('en-IN')}</span>
                  </div>
                  <div className="detail-item" style={{ borderLeft: '1px solid rgba(255,255,255,0.1)', paddingLeft: '2rem' }}>
                    <span className="detail-label">Tenure</span>
                    <span className="detail-value">{offer?.tenure} mo</span>
                  </div>
                  <div className="detail-item" style={{ borderLeft: '1px solid rgba(255,255,255,0.1)', paddingLeft: '2rem' }}>
                    <span className="detail-label">Rate</span>
                    <span className="detail-value">{offer?.interestRate}%</span>
                  </div>
                </div>
                
                <button className="btn-primary" style={{ width: '100%', justifyContent: 'center' }}>
                  <CreditCard size={20} />
                  Accept Offer & Proceed
                </button>
              </>
            )}

          </motion.div>
        )}
      </AnimatePresence>
    </main>
  );
}
