#/usr/bin/env zsh
# start the nix daemon
start_nix_daemon() {
  x=$(pgrep nix-daemon)
  if [ -z "$x" ]; then 
    echo "started new daemon"
    sudo -i --background --non-interactive zsh -c 'nix-daemon >& /tmp/nix-daemon.log'
  else
    echo "nix-daemon running PID=$x"
  fi
}

start_nix_daemon
# pin nixpkgs
echo "downloading and pinnng nixpkgs"
sudo -i nix registry add nixpkgs github:NixOS/nixpkgs
sudo -i nix registry pin nixpkgs
echo "installing base set of packages"
# install some packages
sudo -i nix-env -iA nixpkgs.direnv nixpkgs.starship nixpkgs.git
echo "done"
