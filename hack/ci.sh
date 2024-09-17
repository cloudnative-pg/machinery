#!/usr/bin/env bash

cd "$(dirname "$0")/.." || exit
GITHUB_REF= dagger call -m dagger/ci ci --source .
