#!/bin/bash
set -e
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(dirname "$SCRIPT_DIR")"

# Defaults to the public production swagger. Override with SWAGGER_URL to regenerate
# against a different environment (e.g. a dev-only backend change not yet on prod) —
# never hardcode a non-public URL here, this repo is published publicly.
SWAGGER_URL="${SWAGGER_URL:-https://cloud-api.kvindo.com/swagger/v1/swagger.json}"

echo "Downloading latest API spec from $SWAGGER_URL ..."
curl -f "$SWAGGER_URL" -o "$ROOT_DIR/kvindo-api.json"

echo "Running generator..."
cd "$ROOT_DIR/tools/generator"
go run . --swagger "$ROOT_DIR/kvindo-api.json" --output "$ROOT_DIR/internal/provider"

echo "Building provider..."
cd "$ROOT_DIR"
go build ./...

echo "Done! Provider regenerated successfully."
