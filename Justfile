# This Justfile contains rules/targets/scripts/commands that are used when
# developing. Unlike a Makefile, running `just <cmd>` will always invoke
# that command. For more information, see https://github.com/casey/just
#
#
# this setting will allow passing arguments through to tasks, see the docs here
# https://just.systems/man/en/chapter_24.html#positional-arguments
set positional-arguments

# print all available commands by default
default:
  just --list

# run the test suite
test *args='./...':
  go test "$@"

# lint the entire codebase
lint *args:
  golangci-lint run --fix --config .golangci.yaml "$@"

# build ./cmd/X -> ./bin/X, ./cmd/Y -> ./bin/Y, etc.
build:
  go build -o bin/nix-search ./cmd/nix-search

# builds and pushes peterldowns/nix-search-cli, tagged with :latest and :$COMMIT_SHA
release-container:
  #!/usr/bin/env bash
  COMMIT_SHA=$(git log -1 | head -1 | cut -f 2 -d ' ')
  docker buildx build \
    --platform linux/arm64,linux/amd64,darwin/arm64,darwin/amd64 \
    --label nix-search-cli \
    --tag ghcr.io/peterldowns/nix-search-cli:"$COMMIT_SHA" \
    --tag ghcr.io/peterldowns/nix-search-cli:latest \
    --cache-from ghcr.io/peterldowns/nix-search-cli:latest \
    --build-arg COMMIT_SHA="$COMMIT_SHA" \
    --output type=image,push=true \
    --file ./Dockerfile \
    .

release-binaries:
  #!/usr/bin/env bash
  GOOS=darwin GOARCH=amd64 go build -o ./bin/nix-search-darwin-amd64 ./cmd/nix-search
  GOOS=darwin GOARCH=arm64 go build -o ./bin/nix-search-darwin-arm64 ./cmd/nix-search
  GOOS=linux GOARCH=amd64 go build -o ./bin/nix-search-linux-amd64 ./cmd/nix-search
  GOOS=linux GOARCH=arm64 go build -o ./bin/nix-search-linux-arm64 ./cmd/nix-search
  commit_sha="$(git rev-parse --short HEAD)"
  timestamp="$(date +%s)"
  release_name="release-$timestamp-$commit_sha"
  token="$GITHUB_TOKEN"
  upload_url=$(curl -s -H "Authorization: token $token" \
    -X POST \
    -d "{\"tag_name\": \"$release_name\", \"name\":\"$release_name\",\"target_comitish\": \"$commit_sha\"}" \
    "https://api.github.com/repos/peterldowns/nix-search-cli/releases" | jq -r '.upload_url')
  upload_url="${upload_url%\{*}"
  curl -s -H "Authorization: token $token" \
    -H "Content-Type: application/octet-stream" \
    --data-binary @bin/nix-search-darwin-amd64 \
    "$upload_url?name=nix-search-darwin-amd64&label=nix-search-darwin-amd64"
  curl -s -H "Authorization: token $token" \
    -H "Content-Type: application/octet-stream" \
    --data-binary @bin/nix-search-darwin-arm64 \
    "$upload_url?name=nix-search-darwin-arm64&label=nix-search-darwin-arm64"
  curl -s -H "Authorization: token $token" \
    -H "Content-Type: application/octet-stream" \
    --data-binary @bin/nix-search-linux-amd64 \
    "$upload_url?name=nix-search-linux-amd64&label=nix-search-linux-amd64"
  curl -s -H "Authorization: token $token" \
    -H "Content-Type: application/octet-stream" \
    --data-binary @bin/nix-search-linux-arm64 \
    "$upload_url?name=nix-search-linux-arm64&label=nix-search-linux-arm64"
