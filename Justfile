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
  #!/usr/bin/env sh
  # TODO: allow passing in a specific build target through *args
  for target in $(basename -a ./cmd/*)
  do
    go build -o ./bin/${target} ./cmd/${target}
  done
