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
  };

  outputs = { self, nixpkgs, flake-utils, flake-compat, nix-filter }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
      in
      rec {
        packages = {
          default = pkgs.buildGoModule {
            pname = "demo";
            version = "0.0.1";
            # Every time you update your dependencies (go.mod / go.sum)  you'll
            # need to update the vendorSha256.
            #
            # To find the right hash, set
            #
            #   vendorSha256 = pkgs.lib.fakeSha256;
            #
            # then run `nix build`, take the correct hash from the output, and set
            #
            #   vendorSha256 = <the updated hash>;
            #
            # (Yes, that's really how you're expected to do this.)
            # vendorSha256 = pkgs.lib.fakeSha256;
            vendorSha256 = "sha256-vanKL5s+szW0hduUXGnJNUlyu8wZ2HsBVklIUb/+DLY=";

            src =
              let
                # Set this to `true` in order to show all of the source files
                # that will be included in the module build.
                debug-tracing = false;
                source-files = nix-filter.lib.filter {
                  root = ./.;
                  include = [
                    "./pkg"
                    "./cmd"
                    "go.mod"
                    "go.sum"
                  ];
                };
              in
              (
                if (debug-tracing) then
                  pkgs.lib.sources.trace source-files
                else
                  source-files
              );


            # Add any extra packages required to build the binary should go here.
            buildInputs = [ ];

            # every subpackage will get built with `go build`
            subPackages = [ "cmd/demo" ];
          };
        };
        devShells.default = import ./shell.nix { inherit pkgs; };
      }
    );
}
