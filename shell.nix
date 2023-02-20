{ pkgs ? import <nixpkgs> { } }:
with pkgs;
mkShell {
  buildInputs = with pkgs; [
    delve
    go-outline
    go_1_19
    golangci-lint
    gopkgs
    gopls
    gotools
    just
    nixpkgs-fmt
  ];

  shellHook = ''
    # The path to this repository
    shell_nix="''${IN_LORRI_SHELL:-$(pwd)/shell.nix}"
    workspace_root=$(dirname "$shell_nix")
    export WORKSPACE_ROOT="$workspace_root"

    # We put the $GOPATH/$GOCACHE/$GOENV in $TOOLCHAIN_ROOT,
    # and ensure that the GOPATH's bin dir is on our PATH so tools
    # can be installed with `go install`.
    #
    # Any tools installed explicitly with `go install` will take precedence
    # over versions installed by Nix due to the ordering here.
    export TOOLCHAIN_ROOT="$workspace_root/.toolchain"
    export GOROOT=
    export GOCACHE="$TOOLCHAIN_ROOT/go/cache"
    export GOENV="$TOOLCHAIN_ROOT/go/env"
    export GOPATH="$TOOLCHAIN_ROOT/go/path"
    export GOMODCACHE="$GOPATH/pkg/mod"
    export PATH=$(go env GOPATH)/bin:$PATH
  '';

  # Need to disable fortify hardening because GCC is not built with -oO,
  # which means that if CGO_ENABLED=1 (which it is by default) then the golang
  # debugger fails.
  # see https://github.com/NixOS/nixpkgs/pull/12895/files
  hardeningDisable = [ "fortify" ];
}
