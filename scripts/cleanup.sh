#!/bin/sh
set -eu

docker compose -p tf_go down --volumes --remove-orphans
for path in .cache bin coverage.out; do
  if [ -e "$path" ]; then
    if [ -d "$path" ]; then
      chmod -R u+w "$path"
    fi
    rm -rf -- "$path"
  fi
done
echo "Local tf_go containers, volumes, caches, and binaries removed."
