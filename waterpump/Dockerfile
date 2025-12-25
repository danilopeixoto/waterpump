FROM golang:1.24-alpine3.22 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o waterpump

FROM alpine:3.22

WORKDIR /app

COPY --from=builder /app/waterpump /app/waterpump

CMD ["/app/app"]
