#!/bin/sh
set -eu

if [ "$#" -ne 1 ]; then
  printf '%s\n' "usage: bundle.sh <esper-api-docs-checkout>" >&2
  exit 2
fi

repo=$1
root=$(CDPATH= cd -- "$(dirname -- "$0")/../.." && pwd)
temporary=$(mktemp "${TMPDIR:-/tmp}/esper-openapi.XXXXXX.json")
trap 'rm -f "$temporary"' EXIT

npx --yes @redocly/cli@1.34.5 bundle "$repo/openapi.yaml" --output "$temporary" --ext json
node "$root/tools/specbundle/main.mjs" "$temporary" "$root/spec/openapi"
