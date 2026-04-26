FROM golang:1.24-alpine AS builder
RUN apk add --no-cache protoc
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest && go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
ENV PATH="$PATH:/go/bin"
WORKDIR /app

COPY API-gRPC-Client/go.mod .
COPY API-gRPC-Client/main.go .
COPY Proto/ ./proto/

RUN protoc --go_out=. --go-grpc_out=. proto/report.proto
RUN go mod tidy
RUN go build -o go-deploy-1 main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/go-deploy-1 .
CMD ["./go-deploy-1"]