FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY . .
RUN go mod tidy

RUN CGO_ENABLED=0 GOOS=linux go build -o go-deploy-3 .

FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/go-deploy-3 .

ENV RABBITMQ_URL=amqp://guest:guest@rabbitmq-service.mumnk8s.svc.cluster.local:5672/
ENV VALKEY_URL=redis://valkey-service.mumnk8s.svc.cluster.local:6379/

CMD ["./go-deploy-3"]