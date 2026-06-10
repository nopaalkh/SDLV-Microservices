# Build stage
FROM golang:1.21 as builder

WORKDIR /app
COPY . .

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o e2e-test ./e2e_test_script.go

# Final stage
FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/e2e-test .

CMD ["./e2e-test"]