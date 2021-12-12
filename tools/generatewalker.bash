#!/usr/bin/env bash
set -E -e -o pipefail

grep "^\t*//go:generate go run github.com/newtstat/bqschema-gen-go" -r "$(pwd)" -l | xargs -I{} bash -cx "go generate {}"
