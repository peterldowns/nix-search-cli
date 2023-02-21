#!/usr/bin/env zsh
WORKSPACE_ROOT=/workspaces/$1
echo 'export WORKSPACE_ROOT='$WORKSPACE_ROOT > ~/.bash_aliases
cd $WORKSPACE_ROOT
nix print-dev-env . >> ~/.bash_aliases
