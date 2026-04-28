FROM node:20-alpine

WORKDIR /app

# Install dependencies based on the preferred package manager
COPY package.json package-lock.json* ./
RUN npm ci

COPY src ./src
COPY public ./public
COPY next.config.ts ./
COPY postcss.config.mjs ./
COPY eslint.config.mjs ./
COPY tsconfig.json ./

RUN npm run build

EXPOSE 3000

CMD ["npm", "start"]
