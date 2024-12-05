#!/bin/sh
# scripts/completions.sh

set -e

if ! command -v go >/dev/null; then
  echo '"go" command not found' >&2
  exit 1
fi

scriptdir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)

rm -rf "${scriptdir}/../completions"
mkdir "${scriptdir}/../completions"
for sh in bash fish zsh; do
  go run "${scriptdir}/../main.go" completion "$sh" > "${scriptdir}/../completions/ochami.$sh"
done
