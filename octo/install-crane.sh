#!/usr/bin/env bash

set -euo pipefail

GO111MODULE=on go get -u -ldflags="-s -w" github.com/google/go-containerregistry/cmd/crane
