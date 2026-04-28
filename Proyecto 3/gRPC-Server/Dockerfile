FROM golang:1.24-alpine AS builder
RUN apk add --no-cache protoc
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest && go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
ENV PATH="$PATH:/go/bin"
WORKDIR /app

COPY gRPC-Server/go.mod .
COPY gRPC-Server/main.go .
COPY Proto/ ./proto/

RUN protoc --go_out=. --go-grpc_out=. proto/report.proto

RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -o go-deploy-2 main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/go-deploy-2 .
EXPOSE 50051

ENV RABBITMQ_URL=amqp://guest:guest@rabbitmq-service.mumnk8s.svc.cluster.local:5672/

CMD ["./go-deploy-2"]