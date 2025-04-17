# First stage: build the app
FROM golang:1.24.2-alpine AS builder

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . ./

RUN CGO_ENABLED=0 GOOS=linux go build -o main cmd/main.go

# Second stage: minimal & secure runtime container with distroless (from google) 
FROM gcr.io/distroless/base

WORKDIR /usr/local/bin

COPY --from=builder /app/main .

EXPOSE 8443

CMD ["./main"]
