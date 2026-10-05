# syntax=docker/dockerfile:1
ARG GO_VERSION=1.27

FROM golang:${GO_VERSION}-alpine AS build
WORKDIR /src
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/encoder ./cmd/encoder

FROM alpine:3.20
RUN apk add --no-cache ffmpeg ca-certificates \
 && adduser -D -u 1000 app
USER app
WORKDIR /app
COPY --from=build /out/encoder /usr/local/bin/encoder
ENV INPUT_DIR=/data/input OUTPUT_DIR=/data/output
ENTRYPOINT ["encoder"]
