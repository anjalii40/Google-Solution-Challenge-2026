FROM golang:1.21-alpine

WORKDIR /app

# Enable automatic go module downloads
ENV GO111MODULE=on

COPY go.mod ./
# Copy go.sum if it exists; for now we'll just run go mod tidy
RUN go mod tidy || true

COPY . .

RUN go mod tidy
RUN go build -o main .

EXPOSE 8080

CMD ["./main"]
