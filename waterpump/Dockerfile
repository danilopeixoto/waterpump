FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o waterpump

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/waterpump /app/waterpump

CMD ["/app/app"]
