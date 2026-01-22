# -------------- builder container --------------
FROM golang:1.25.5 as builder

WORKDIR /app

ARG VERSION=latest
ARG COMMIT=HEAD

# Copy go mod files first for better caching
COPY go.mod go.sum ./

RUN go env -w GO111MODULE=on
RUN go mod download

# Install build tools in a separate cached layer
RUN mkdir -p /app/bin
RUN GOBIN=/app/bin go install github.com/swaggo/swag/v2/cmd/swag@latest

COPY . .

RUN make build -e VERSION=${VERSION} -e GITCOMMIT=${COMMIT}

# -------------- runner container --------------
FROM alpine:3.19 AS runner
WORKDIR /app

COPY --from=builder /app/apiserver /usr/bin/apiserver


# docs
RUN mkdir -p /app/docs

COPY --from=builder /app/docs /app/docs

ENV DOC_FILE_BASE_DIR=/app/docs

# logs
RUN mkdir -p /app/v3logs

ENV LOG_FILE_BASE_DIR=/app/v3logs

