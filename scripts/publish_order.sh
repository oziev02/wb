#!/usr/bin/env bash
set -euo pipefail
BROKERS="${BROKERS:-localhost:9092}"
TOPIC="${TOPIC:-orders}"
MSG_FILE="${1:-scripts/seed_order.json}"

if ! command -v kafka-console-producer >/dev/null 2>&1; then
  echo "Need kafka-console-producer in PATH" >&2
  exit 1
fi

cat "$MSG_FILE" | kafka-console-producer --broker-list "$BROKERS" --topic "$TOPIC"
