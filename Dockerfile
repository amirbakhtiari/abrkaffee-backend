FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /abrcoffee cmd/api/main.go

FROM alpine:3.19
WORKDIR /app

RUN apk --no-cache add tzdata

COPY --from=builder /abrcoffee .
EXPOSE 8080
CMD ["./abrcoffee"]
