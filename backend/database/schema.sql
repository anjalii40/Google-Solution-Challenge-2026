CREATE TABLE IF NOT EXISTS sessions (
    session_id UUID PRIMARY KEY,
    status VARCHAR(50) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS session_logs (
    id SERIAL PRIMARY KEY,
    session_id UUID REFERENCES sessions(session_id),
    log_message TEXT NOT NULL,
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS transcripts (
    id SERIAL PRIMARY KEY,
    session_id UUID REFERENCES sessions(session_id),
    name VARCHAR(255),
    declared_income INT,
    loan_purpose TEXT,
    employment VARCHAR(100),
    verbal_consent BOOLEAN,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS cv_results (
    id SERIAL PRIMARY KEY,
    session_id UUID REFERENCES sessions(session_id),
    estimated_age_min INT,
    estimated_age_max INT,
    declared_age INT,
    flag BOOLEAN,
    flag_reason VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS geo_results (
    id SERIAL PRIMARY KEY,
    session_id UUID REFERENCES sessions(session_id),
    ip_location VARCHAR(255),
    gps_location VARCHAR(255),
    location_mismatch BOOLEAN,
    vpn_detected BOOLEAN,
    device VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS llm_outputs (
    id SERIAL PRIMARY KEY,
    session_id UUID REFERENCES sessions(session_id),
    risk_band VARCHAR(50),
    flags JSONB,
    recommendation VARCHAR(50),
    confidence FLOAT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS loan_offers (
    id SERIAL PRIMARY KEY,
    session_id UUID REFERENCES sessions(session_id),
    status VARCHAR(50),
    reason TEXT,
    amount INT,
    emi INT,
    tenure INT,
    interest_rate FLOAT,
    risk_tier VARCHAR(50),
    flags JSONB,
    manual_review_required BOOLEAN,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
