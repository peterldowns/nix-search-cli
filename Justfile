# This Justfile contains rules/targets/scripts/commands that are used when
# developing. Unlike a Makefile, running `just <cmd>` will always invoke
# that command. For more information, see https://github.com/casey/just
#
#
# this setting will allow passing arguments through to tasks, see the docs here
# https://just.systems/man/en/chapter_24.html#positional-arguments
set positional-arguments

# print all available commands by default
help:
  @just --list

# run the test suite
test *args='./...':
  go test "$@"

# lint the entire codebase
lint *args:
  golangci-lint run --fix --config .golangci.yaml "$@"

# build ./bin/nix-search
build:
  #!/usr/bin/env bash
  ldflags=$(./scripts/golang-ldflags.sh)
  go build -ldflags "$ldflags" -o bin/nix-search ./cmd/nix-search

# build ghcr.io/peterldowns/nix-search-cli:local
#
# run with
#    docker run -it --rm ghcr.io/peterldowns/nix-search-cli:local
build-docker:
  #!/usr/bin/env bash
  COMMIT_SHA=$(git rev-parse --short HEAD || echo "unknown")
  VERSION=$(cat ./VERSION)
  docker build \
    --tag ghcr.io/peterldowns/nix-search-cli:local \
    --build-arg COMMIT_SHA="$COMMIT_SHA" \
    --build-arg VERSION="$VERSION" \
    --file ./Dockerfile \
    -o type=image \
    .
