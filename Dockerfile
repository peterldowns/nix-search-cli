# Build Stage
# Pin to the latest minor version without specifying a patch version so that
# we always deploy security fixes as soon as they are available.
FROM golang:1.22-alpine as builder

# Have to put our source in the right place for it to build
WORKDIR $GOPATH/src/github.com/peterldowns/nix-search-cli

ENV GO111MODULE=on
ENV CGO_ENABLED=0

# Install the dependencies
COPY go.mod .
COPY go.sum .
RUN go mod download

# Build the application
COPY . .

# Put the appropriate build artifacts in a folder for distribution
RUN mkdir -p /dist

ARG VERSION=unknown
ARG COMMIT_SHA=unknown
ENV NSC_VERSION=$VERSION
ENV NSC_COMMIT_SHA=$COMMIT_SHA
RUN go build \
  -ldflags "-X main.Version=${NSC_VERSION} -X main.Commit=${NSC_COMMIT_SHA}" \
  -o /dist/nix-search \
  ./cmd/nix-search

# App Stage
FROM alpine:3.20.3 as app

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

ARG VERSION=unknown
ARG COMMIT_SHA=unknown
ENV NSC_VERSION=$VERSION
ENV NSC_COMMIT_SHA=$COMMIT_SHA

LABEL org.opencontainers.image.source="https://github.com/peterldowns/nix-search-cli"
LABEL org.opencontainers.image.description="nix-search"
LABEL org.opencontainers.image.licenses="MIT"
LABEL org.opencontainers.image.version="${NSC_VERSION}"
LABEL org.opencontainers.image.revision="${NSC_COMMIT_SHA}"


ARG COMMIT_SHA=null
ENV COMMIT_SHA=$COMMIT_SHA

COPY --from=builder --chown=app:app /dist /app
ENV PATH="/app:$PATH"
# override and get a shell with `docker run --entrypoint=ash`
ENTRYPOINT ["/app/nix-search"]
