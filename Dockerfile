FROM golang:1.23.6 AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# build the go binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o app .


FROM alpine:latest
WORKDIR /root/

# Install CA certificates (needed for HTTPS requests)
RUN apk add --no-cache ca-certificates

COPY --from=builder /app/app .

CMD ["./app"]