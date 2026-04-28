#!/bin/bash
set -e

echo "Deploying GDG Solution Project to Google Cloud Run..."

PROJECT_ID=$(gcloud config get-value project)
REGION="asia-south1"

# Deploy Go Backend
echo "Deploying Backend (Go WebSockets)..."
gcloud run deploy gdg-backend \
  --source ./backend \
  --platform managed \
  --region $REGION \
  --allow-unauthenticated \
  --set-env-vars DB_HOST="<YOUR-DATABASE-URL>",DB_USER="admin",DB_PASSWORD="password",DB_NAME="onboard_db",DB_PORT="5432"

# Retrieve Backend URL
BACKEND_URL=$(gcloud run services describe gdg-backend --platform managed --region $REGION --format 'value(status.url)' | sed 's/http\(s*\):\/\//ws\1:\/\//')

# Deploy Next.js Frontend
echo "Deploying Frontend (Next.js)..."
gcloud run deploy gdg-frontend \
  --source . \
  --platform managed \
  --region $REGION \
  --allow-unauthenticated \
  --set-env-vars NEXT_PUBLIC_WS_URL="$BACKEND_URL/ws/onboard"

echo "Deployment completed successfully! Run 'gcloud run services list' to see live endpoints."
