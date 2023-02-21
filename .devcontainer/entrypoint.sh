#!/usr/bin/env zsh
start_nix_daemon() {
  x=$(pgrep nix-daemon)
  if [[ -z "$x" ]]; then 
    echo "started new daemon"
    sudo -i --background --non-interactive zsh -c 'nix-daemon >& /tmp/nix-daemon.log'
  else
    echo "nix-daemon running PID=$x"
  fi
}
start_nix_daemon
exec "$@"
