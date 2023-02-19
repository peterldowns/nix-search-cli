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
release:
  #!/usr/bin/env bash
  COMMIT_SHA=$(git log -1 | head -1 | cut -f 2 -d ' ')
  docker buildx build \
    --platform linux/arm64,linux/amd64 \
    --label migrate \
    --tag ghcr.io/peterldowns/nix-search-cli:"$COMMIT_SHA" \
    --tag ghcr.io/peterldowns/nix-search-cli:latest \
    --cache-from ghcr.io/peterldowns/nix-search-cli:latest \
    --build-arg COMMIT_SHA="$COMMIT_SHA" \
    --output type=image,push=true \
    --file ./Dockerfile \
    .
