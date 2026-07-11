#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CONSOLE_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"

cd "${CONSOLE_DIR}"
mkdir -p .build/front .build/backend

docker buildx build \
  --file .docker/front.Dockerfile \
  --target artifacts \
  --output type=local,dest=.build/front \
  .

docker buildx build \
  --file .docker/backend.Dockerfile \
  --target artifacts \
  --output type=local,dest=.build/backend \
  .

docker compose -f .docker/docker-compose.yml build front backend
