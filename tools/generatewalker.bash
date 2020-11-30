#!/usr/bin/env bash
set -E -e -o pipefail

# NOTE(djeeno): https://golang.org/cmd/go/#hdr-Generate_Go_files_by_processing_source
#               To convey to humans and machine tools that code is generated, generated source should have a line that matches the following regular expression (in Go syntax):
#                   ^// Code generated .* DO NOT EDIT\.$
check_string="by go run github.com/djeeno/bqtableschema;"
regex="^// Code generated ${check_string:?} DO NOT EDIT\.$"

grep "${regex:?}" -r "$(pwd)" -l | xargs -I{} bash -cx "go generate {}"
