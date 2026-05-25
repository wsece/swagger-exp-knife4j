# syntax=docker/dockerfile:1

FROM golang:1.26-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath \
    -ldflags="-s -w" \
    -o /out/swagger-exp-knife4j .

FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata wget \
    && adduser -D -u 1000 app

WORKDIR /data

COPY --from=builder /out/swagger-exp-knife4j /usr/local/bin/swagger-exp-knife4j

RUN mkdir -p /data/output && chown -R app:app /data

USER app

VOLUME ["/data"]

EXPOSE 7171

# 数据目录：/data/swagger-scan.sqlite3、/data/output/
ENTRYPOINT ["swagger-exp-knife4j"]
CMD ["report", "server", "--host", "0.0.0.0", "--port", "7171", "--db-uri", "sqlite://swagger-scan.sqlite3", "--api-doc-path", "./output"]
