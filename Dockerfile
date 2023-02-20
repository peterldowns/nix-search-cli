# Build Stage
# Pin to the latest minor version without specifying a patch version so that
# we always deploy security fixes as soon as they are available.
FROM golang:1.18-alpine as builder

# Have to put our source in the right place for it to build
WORKDIR $GOPATH/src/github.com/peterldowns/nix-search-cli

ENV GO111MODULE=on
ENV CGO_ENABLED=1

# Install the dependencies
COPY go.mod .
COPY go.sum .
RUN go mod download

# Build the application
COPY . .

# Put the appropriate build artifacts in a folder for distribution
RUN mkdir -p /dist
RUN go build -o /dist/nix-search ./cmd/nix-search

# App Stage
FROM alpine:3.16.3 as app
LABEL org.opencontainers.image.source="https://github.com/peterldowns/nix-search-cli"
LABEL org.opencontainers.image.description="nix-search"
LABEL org.opencontainers.image.licenses="MIT"

# Add a non-root user and group with appropriate permissions
# and consistent ids.
RUN addgroup --gid 888 --system app && \
    adduser --no-create-home \
            --gecos "" \
            --shell "/bin/ash" \
            --uid 999 \
            --ingroup app \
            --system \
            app
USER app
WORKDIR /app

ARG COMMIT_SHA=null
ENV COMMIT_SHA=$COMMIT_SHA

COPY --from=builder --chown=app:app /dist /app
ENTRYPOINT ["/app/nix-search"]
