#!/usr/bin/env bash
set -euo pipefail

upstream_url="$1"
git fetch --quiet --no-tags "$upstream_url" '+refs/tags/v*:refs/tags/v*'

base="$(git tag --merged HEAD --list 'v*' | grep -E '^v[0-9]+\.[0-9]+\.[0-9]+$' | sort -V | tail -n 1 || true)"
if [[ -z "$base" ]]; then
  echo "no upstream release tag is reachable from HEAD" >&2
  exit 1
fi

last="$(git tag --list "${base}-clepsydra.*" | sed -nE "s/^${base//./\\.}-clepsydra\.([0-9]+)$/\1/p" | sort -n | tail -n 1)"
echo "${base#v}-clepsydra.$((${last:-0} + 1))"
