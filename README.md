# nix-search-cli

A CLI client for the [`search.nixos.org/packages index`](https://search.nixos.org/packages).
Use `nix-search` to find packages by name, description, installed programs, or other metadata.
Does not work offline.

```bash
# Search for a package
nix-search <text to match>

# Use a specific channel
nix-search --channel unstable --query <text to match>
# Show full usage / help
nix-search --help
```

For example, figuring out how to install `gcloud`:
```shell
nix-search gcloud
```
```
google-cloud-sdk-gce -> [bq, docker-credential-gcloud, gcloud, gsutil, git-credential-gcloud.sh]
google-cloud-sdk -> [git-credential-gcloud.sh, docker-credential-gcloud, gcloud, bq, gsutil]
rPackages.tagcloud
perl536Packages.HTMLTagCloud
perl534Packages.HTMLTagCloud
```

## Install

Golang:

```bash
go install github.com/peterldowns/nix-search-cli
```

## Motivation
Nix is useful as a way to install packages, but without this project there is no easy way to find the attribute name
to use to install a given program.

The [Nix Wiki page on "Searching Packages"](https://nixos.wiki/wiki/Searching_packages) recommends
using the `search.nixos.org` interface, but doing this requires using a browser.

As for `nix-env --query`, it supports searching over attribute names, but not
other fields or metadata (including the programs that the attribute installs).

For instance, you can use `nix-env -qaP` to search for
attribute names:

```bash
# nix-env -qaP google-cloud-sdk
nixpkgs.google-cloud-sdk      google-cloud-sdk-408.0.1
nixpkgs.google-cloud-sdk-gce  google-cloud-sdk-408.0.1
```

but you cannot find an attribute name given a binary you'd like to install:

```bash
# nix-env -qaP gcloud
error: selector 'gcloud' matches no derivations
```

## Contributing

Common tasks are run by `just`
```bash
# show all available commands
just
just --list
```

This repository is compatible with nix (standard), nix (flakes), direnv, and
lorri. You can explicitly enter a development shell with all necessary
dependencies with either `nix develop` (flakes) or `nix shell` (standard).

This repository ships configuration details for VSCode. After entering a
development shell, run `code .` from the root of the repository to open VSCode.

```bash
# get developer dependencies by entering a nix shell.
# if you have direnv / lorri installed, you just need to allow the config once.
nix develop # (flakes)
nix-shell # (standard)
direnv allow # direnv
```

### Testing and Linting:
```bash
just test
just lint
```

### Building:
```bash
# build with `go build`, result is in `./bin/demo`
just build
# build with `nix`, result is in `./result/bin/demo`
nix build # (flakes)
nix-build # (standard)
```

### Run the binary:
```bash
# if built with `just build`:
./bin/demo help
# if built with `nix build` or `nix-build`:
./result/bin/demo help
# or, you can build + run directly through nix:
nix run . -- help # flakes
# or, you can open a new shell with the binary available on $PATH through nix:
nix shell # and then `nix-search`
nix shell -c demo help # directly run `nix-search` from inside this shell
```

### Update the flake.lock:
```bash
# Re-generate the flake.lock file
nix flake lock
# Update all dependencies and update the flake.lock file
nix flake update
```