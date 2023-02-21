# nix-search-cli

`nix-search` is a CLI client for [`search.nixos.org/packages`](https://search.nixos.org/packages).
Use `nix-search` to find packages by name, description, installed programs, version, or other metadata. Requires an active internet connection to work.

Major features and benefits:
* Find how to install the binary you need
* Searches work the same as the web interface by default
* Use flags to explicitly query attribute names, installed programs, and versions
* Each result is linked to the web interface (in supported terminals)
* Results are compact and nicely colorized by default (in supported terminals)

```console
$ nix-search --help
search for packages via search.nixos.org

Usage:
  nix-search ...query [flags]

Examples:
  # Search
  # ... like the web interface
  nix-search python linter
  nix-search --search "python linter"
  # ... by attr name
  nix-search --attr python
  nix-search --attr 'emacsPackages.*'
  # ... by version
  nix-search --version 1.20
  nix-search --version '1.*'
  # ... by installed programs
  nix-search --program python
  nix-search --program "py*"
  # ... with ElasticSearch QueryString syntax
  nix-search --query-string="package_programs:(crystal OR irb)"
  nix-search --query-string='package_description:(MIT Scheme)'
  # ... with multiple filters and options
  nix-search --attr go --version 1.20 --details

Flags:
  -a, --attr string           search by attr name
  -c, --channel string        which channel to search in (default "unstable")
  -d, --details               show expanded details for each result
  -h, --help                  help for nix-search
  -j, --json                  emit results in json-line format
  -m, --max-results int       maximum number of results to return (default 20)
  -p, --program string        search by installed programs
  -q, --query-string string   perform an advanced query string format search
  -s, --search string         default search, same as the website
  -v, --version string        search by version
```

For example, here's how you would find all packages that install a `gcloud` binary. The results show the version of each package as well as the full set of installed binaries. In a supported terminal, we use nice colors:

```console
$ ./bin/nix-search -p gcloud
google-cloud-sdk-gce @ 408.0.1: gcloud bq docker-credential-gcloud git-credential-gcloud.sh gsutil
google-cloud-sdk @ 408.0.1: gcloud bq docker-credential-gcloud git-credential-gcloud.sh gsutil
```

Here's how you would find out how to install python 3.12:

```console
$ ./bin/nix-search -p python -v 3.12
python312 @ 3.12.0a5: python 2to3 2to3-3.12 idle idle3 idle3.12 pydoc pydoc3 pydoc3.12 python-config python3 python3-config python3.12 python3.12-config
```

## Install

Golang:
```bash
# run it
go run github.com/peterldowns/nix-search-cli/cmd/nix-search@latest --help
# install it
go install github.com/peterldowns/nix-search-cli/cmd/nix-search@latest
```

Docker:
```bash
# run it
docker run --rm -it ghcr.io/peterldowns/nix-search-cli:latest --help
# pull it
docker pull ghcr.io/peterldowns/nix-search-cli:latest
```

Nix (flakes):
```bash
# run it
nix run github:peterldowns/nix-search-cli --help
# install it
nix profile install github:peterldowns/nix-search-cli --refresh
```

Manual:
- Visit [the latest Github release](https://github.com/peterldowns/nix-search-cli/releases/latest)
- Download the appropriate binary: `nix-search-$os-$arch`

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
# build with `go build`, result is in `./bin/nix-search`
just build
# build with `nix`, result is in `./result/bin/nix-search`
nix build # (flakes)
nix-build # (standard)
```

### Run the binary:
```bash
# if built with `just build`:
./bin/nix-search --help
# if built with `nix build` or `nix-build`:
./result/bin/nix-search --help
# or, you can build + run directly through nix:
nix run . -- help # flakes
# or, you can open a new shell with the binary available on $PATH through nix:
nix shell # and then `nix-search`
nix shell -c nix-search --help # directly run `nix-search` from inside this shell
```

### Updating the gomod2nix file
If you make changes that modify the golang dependencies, you'll need to update the pinned dependencies used in the Nix build process:

```bash
gomod2nix
```

### Update the flake.lock:
```bash
# Re-generate the flake.lock file
nix flake lock
# Update all dependencies and update the flake.lock file
nix flake update
```
