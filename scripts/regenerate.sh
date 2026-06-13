#!/bin/bash
set -e
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(dirname "$SCRIPT_DIR")"

echo "Downloading latest API spec..."
curl -f "https://cloud-api.kvindo.com/swagger/v1/swagger.json" -o "$ROOT_DIR/kvindo-api.json"

echo "Running generator..."
cd "$ROOT_DIR/tools/generator"
go run . --swagger "$ROOT_DIR/kvindo-api.json" --output "$ROOT_DIR/internal/provider"

echo "Building provider..."
cd "$ROOT_DIR"
go build ./...

echo "Done! Provider regenerated successfully."
