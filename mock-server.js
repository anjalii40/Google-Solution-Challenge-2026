const express = require('express');
const http = require('http');
const { Server } = require('socket.io');

const app = express();
const server = http.createServer(app);
const io = new Server(server, {
  cors: {
    origin: '*',
    methods: ['GET', 'POST']
  }
});

io.on('connection', (socket) => {
  console.log('Client connected:', socket.id);

  let frameCount = 0;
  let audioChunkCount = 0;
  let verificationTimer;

  // Listen for audio chunks
  socket.on('audio_chunk', (data) => {
    audioChunkCount++;
    console.log(`[${socket.id}] Received audio chunk ${audioChunkCount}. Size: ${data.length} bytes`);
  });

  // Listen for video frames
  socket.on('video_frame', (data) => {
    frameCount++;
    console.log(`[${socket.id}] Received video frame ${frameCount}. Payload preview: ${data.substring(0, 30)}...`);
    
    // Simulate backend processing and sending statuses back
    if (frameCount === 1) {
      socket.emit('status_update', { message: 'Verifying identity...' });
    } else if (frameCount === 3) {
      socket.emit('status_update', { message: 'Analyzing risk profile...' });
    } else if (frameCount === 5) {
      socket.emit('status_update', { message: 'Generating tailored offer...' });
      
      // Simulate Final Offer
      verificationTimer = setTimeout(() => {
        const offer = {
          amount: 500000,
          emi: 12500,
          tenure: 48,
          interestRate: 10.5
        };
        console.log(`[${socket.id}] Sending offer to client:`, offer);
        socket.emit('offer_received', offer);
      }, 3000);
    }
  });

  socket.on('disconnect', () => {
    console.log('Client disconnected:', socket.id);
    if (verificationTimer) clearTimeout(verificationTimer);
  });
});

const PORT = 4000;
server.listen(PORT, () => {
  console.log(`Mock WebSocket Server running on http://localhost:${PORT}`);
});
