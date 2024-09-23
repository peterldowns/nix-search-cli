{
  description = "CLI for searching packages on search.nixos.org";
  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs";

    flake-utils.url = "github:numtide/flake-utils";

    flake-compat.url = "github:edolstra/flake-compat";
    flake-compat.flake = false;

    nix-filter.url = "github:numtide/nix-filter";
  };

  outputs = { self, ... }@inputs:
    inputs.flake-utils.lib.eachDefaultSystem (system:
      let
        overlays = [ ];
        pkgs = import inputs.nixpkgs {
          inherit system overlays;
        };
        version = (builtins.readFile ./VERSION);
        commit = if (builtins.hasAttr "rev" self) then (builtins.substring 0 7 self.rev) else "unknown";
      in
      rec {
        packages = rec {
          nix-search = pkgs.buildGoModule {
            pname = "nix-search";
            version = version;
            # Every time you update your dependencies (go.mod / go.sum)  you'll
            # need to update the vendorSha256.
            #
            # To find the right hash, set
            #
            #   vendorHash = pkgs.lib.fakeHash;
            #
            # then run `nix build`, take the correct hash from the output, and set
            #
            #   vendorHash = <the updated hash>;
            #
            # (Yes, that's really how you're expected to do this.)
            #vendorHash = pkgs.lib.fakeHash;
            vendorHash = "sha256-RZuB0aRiMSccPhX30cGKBBEMCSvmC6r53dWaqDYbmyA=";
            src =
              let
                # Set this to `true` in order to show all of the source files
                # that will be included in the module build.
                debug-tracing = false;
                source-files = inputs.nix-filter.lib.filter {
                  root = ./.;
                };
              in
              (
                if (debug-tracing) then
                  pkgs.lib.sources.trace source-files
                else
                  source-files
              );

            GOWORK = "off";
            modRoot = ".";
            subPackages = [
              "cmd/nix-search"
            ];
            ldflags = [
              "-X main.Version=${version}"
              "-X main.Commit=${commit}"
            ];

            # Add any extra packages required to build the binaries should go here.
            buildInputs = [ ];
            doCheck = false;
          };
          default = nix-search;
        };

        apps = rec {
          nix-search = {
            type = "app";
            program = "${packages.nix-search}/bin/nix-search";
          };
          default = nix-search;
        };

        devShells = rec {
          default = pkgs.mkShell {
            packages = with pkgs; [
              ## golang
              delve
              go-outline
              go
              golangci-lint
              gopkgs
              gopls
              gotools
              ## nix
              nixpkgs-fmt
              ## other tools
              just
            ];

            shellHook = ''
              # The path to this repository
              if [ -z $WORKSPACE_ROOT ]; then
                shell_nix="''${IN_LORRI_SHELL:-$(pwd)/shell.nix}"
                workspace_root=$(dirname "$shell_nix")
                export WORKSPACE_ROOT="$workspace_root"
              fi

              # We put the $GOPATH/$GOCACHE/$GOENV in $TOOLCHAIN_ROOT,
              # and ensure that the GOPATH's bin dir is on our PATH so tools
              # can be installed with `go install`.
              #
              # Any tools installed explicitly with `go install` will take precedence
              # over versions installed by Nix due to the ordering here.
              export TOOLCHAIN_ROOT="$WORKSPACE_ROOT/.toolchain"
              export GOROOT=
              export GOCACHE="$TOOLCHAIN_ROOT/go/cache"
              export GOENV="$TOOLCHAIN_ROOT/go/env"
              export GOPATH="$TOOLCHAIN_ROOT/go/path"
              export GOMODCACHE="$GOPATH/pkg/mod"
              export PATH=$(go env GOPATH)/bin:$PATH
              # This project is pure go and does not need CGO. We disable it
              # here as well as in the Dockerfile and nix build scripts.
              export CGO_ENABLED=0
            '';

            # Need to disable fortify hardening because GCC is not built with -oO,
            # which means that if CGO_ENABLED=1 (which it is by default) then the golang
            # debugger fails.
            # see https://github.com/NixOS/nixpkgs/pull/12895/files
            hardeningDisable = [ "fortify" ];
          };
        };
      }
    );
}
