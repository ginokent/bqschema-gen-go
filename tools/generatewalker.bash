#!/usr/bin/env bash
set -E -e -o pipefail

grep "^\t*//go:generate go run github.com/djeeno/bqschema-gen-go/v0" -r "$(pwd)" -l | xargs -I{} bash -cx "go generate {}"
