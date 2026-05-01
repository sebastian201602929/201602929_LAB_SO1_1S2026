FROM rust:latest AS builder
WORKDIR /app
COPY Cargo.toml Cargo.lock ./
COPY src ./src
RUN cargo build --release

FROM debian:bookworm-slim
WORKDIR /app
RUN apt-get update && apt-get install -y libssl3
COPY --from=builder /app/target/release/api_rust /usr/local/bin/api_rust
EXPOSE 8080

ENV GRPC_CLIENT_URL=http://grpc-client-service.mumnk8s.svc.cluster.local:8081/grpc-201602929

CMD ["api_rust"]