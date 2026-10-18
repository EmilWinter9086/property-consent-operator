#!/bin/sh
set -eu

: "${INFRAI_API_KEY:?set INFRAI_API_KEY first}"

go run . &
service_pid=$!
trap 'kill "$service_pid" 2>/dev/null || true' EXIT

sleep 1
curl --fail-with-body --silent --show-error \
  -X POST http://127.0.0.1:8080/consent \
  -H 'Content-Type: application/json' \
  -d '{"user_id":"tenant-42","resource":"inspection_reminder","action":"grant","operation_id":"inspection-opt-in-2026-09"}'
printf '\n'
