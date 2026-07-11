#!/usr/bin/env bash
set -euo pipefail

/app/gateway-controller &
gateway_controller_pid=$!

/app/gateway &
gateway_pid=$!

terminate() {
	kill "${gateway_controller_pid}" "${gateway_pid}" 2>/dev/null || true
	wait "${gateway_controller_pid}" 2>/dev/null || true
	wait "${gateway_pid}" 2>/dev/null || true
}

trap terminate INT TERM

if wait -n "${gateway_controller_pid}" "${gateway_pid}"; then
	status=0
else
	status=$?
	terminate
	exit "${status}"
fi

terminate
