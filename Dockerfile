FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/app

FROM alpine:latest

WORKDIR /app

RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

# cerate logs directory
RUN mkdir -p /var/log/li-acc && chown appuser:appuser /var/log/li-acc

COPY --from=builder /app/server .

COPY --from=builder /app/.env.example ./.env.example
COPY --from=builder /app/.env ./.env

COPY --from=builder /app/internal/repository/db/migrations ./internal/repository/db/migrations

COPY --from=builder /app/templates ./templates
COPY --from=builder /app/static ./static
COPY --from=builder /app/assets/excel/blank_receipt_pattern.xls ./assets/excel/blank_receipt_pattern.xls
COPY --from=builder /app/static/fonts/Arial.ttf ./static/fonts/Arial.ttf


EXPOSE 8080

CMD ["./server"]

