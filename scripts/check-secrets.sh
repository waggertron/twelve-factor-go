#!/bin/sh
set -eu

if git grep -n -E 'sk_live_|ghp_|AKIA|xoxb-|eyJ[A-Za-z0-9_-]' -- ':!go.sum' ':!scripts/check-secrets.sh'; then
  echo "credential-shaped content found" >&2
  exit 1
fi
echo "Secret-pattern scan passed."
