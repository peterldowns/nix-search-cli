#!/bin/bash

start_nix_daemon() {
  x=$(pgrep nix-daemon)
  if [[ -z "$x" ]]; then
    sudo -i --background --non-interactive zsh -c 'nix-daemon >& /tmp/nix-daemon.log'
  fi
}

start_nix_daemon
BASH_ENV=~/.bash_aliases bash "$@"
