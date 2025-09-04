#!/usr/bin/env bash

set -e

export DEBIAN_FRONTEND=noninteractive

# Checks if packages are installed and installs them if not
check_packages() {
  if ! dpkg -s "$@" > /dev/null 2>&1; then
    if [ "$(find /var/lib/apt/lists/* | wc -l)" = "0" ]; then
      echo "Running apt-get update..."
      apt-get update -y
    fi
    apt-get -y install --no-install-recommends "$@"
  fi
}

clean_up() {
  apt-get autoremove -y
  apt-get clean -y
  rm -rf /var/lib/apt/lists/*
}

# The ghcr.io/guiyomh/features/just:0 feature does not install the fzf command, so the `just --choose` command cannot be used.
check_packages fzf

clean_up