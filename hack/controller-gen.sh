#!/usr/bin/env bash

cd "$(dirname "$0")/.." || exit
GITHUB_REF= dagger call -m dagger/ci controller-gen --source . \
    file --path pkg/api/zz_generated.deepcopy.go  \
    export --path pkg/api/zz_generated.deepcopy.go
