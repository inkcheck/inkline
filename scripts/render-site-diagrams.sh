#!/bin/sh
# Renders every example in both themes into the site's static files.
set -eu
cd "$(dirname "$0")/.."
go build -o bin/inkline ./cmd/inkline
out=site/static/diagrams
mkdir -p "$out"
for f in examples/*.d2; do
  name=$(basename "$f" .d2)
  for theme in light dark; do
    ./bin/inkline render "$f" --theme "$theme" -o "$out/$name-$theme.svg"
  done
done
# The kit pages show examples in their own kit.
for name in architecture flowchart; do
  for theme in light dark; do
    ./bin/inkline render "examples/$name.d2" --design-system fluent --theme "$theme" -o "$out/$name-fluent-$theme.svg"
  done
done
echo "rendered $(ls "$out" | wc -l | tr -d ' ') diagrams into $out"
