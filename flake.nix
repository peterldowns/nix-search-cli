{
  description = "demo is a golang binary";
  inputs = {
    nixpkgs = {
      url = "github:nixos/nixpkgs/nixos-unstable";
    };
    flake-utils = {
      url = "github:numtide/flake-utils";
    };
    flake-compat = {
      url = "github:edolstra/flake-compat";
      flake = false;
    };
    nix-filter = {
      url = github:numtide/nix-filter;
    };
    gomod2nix = {
      url = "github:nix-community/gomod2nix";
    };
  };

  outputs = { self, nixpkgs, flake-utils, flake-compat, nix-filter, gomod2nix }:
    flake-utils.lib.eachDefaultSystem
      (system:
        let
          pkgs = import nixpkgs {
            inherit system;
            overlays = [ gomod2nix.overlays.default ];
          };
        in
        rec {
          packages = rec {
            nix-search = pkgs.buildGoApplication {
              pname = "nix-search-cli";
              version = "0.0.1";
              src = ./.;
              modules = ./gomod2nix.toml;
            };
            # Makes `nix build` == `nix build .#nix-search`
            default = nix-search;
          };


          apps = rec {
            nix-search = {
              type = "app";
              program = "${packages.nix-search}/bin/nix-search";
            };
            # Makes `nix run .#nix-search` work.
            default = nix-search;
          };

          devShells.default = import ./shell.nix {
            inherit pkgs;
          };
        }
      );
}
