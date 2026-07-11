#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CONSOLE_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"

cd "${CONSOLE_DIR}"
docker compose -f .docker/docker-compose.yml up -d front backend
