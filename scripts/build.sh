#!/usr/bin/env bash
set -exuo pipefail

cd "$( dirname "${BASH_SOURCE[0]}" )/.."
source .envrc

GOOS=linux go build -ldflags="-s -w" -o bin/supply ca-ncore/supply/cli
GOOS=windows go build -ldflags="-s -w" -o bin/supply.exe ca-ncore/supply/cli
