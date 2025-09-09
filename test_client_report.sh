#!/bin/bash

# Пример тестирования client report endpoint'а
# Используйте этот скрипт для проверки функциональности

echo "Testing client report handling..."

# Пример client report payload как в вашем issue
CLIENT_REPORT_PAYLOAD=$'{}
{"type":"client_report"}
{"timestamp":1740239731.564,"discarded_events":[{"reason":"network_error","category":"internal","quantity":1}]}'

echo "Sending client report to endpoint..."
echo "$CLIENT_REPORT_PAYLOAD"

# Замените PROJECT_ID и SENTRY_KEY на реальные значения
PROJECT_ID="1"
SENTRY_KEY="dfe8c27ddfdd46345d60a348f2896334"

echo ""
echo "POST /api/${PROJECT_ID}/envelope/?sentry_key=${SENTRY_KEY}&sentry_version=7&sentry_client=sentry.javascript.node%2F8.26.0"
echo ""

# Раскомментируйте для реального тестирования:
# curl -v \
#   -X POST \
#   -H "Content-Type: text/plain" \
#   -d "$CLIENT_REPORT_PAYLOAD" \
#   "http://localhost:8080/api/${PROJECT_ID}/envelope/?sentry_key=${SENTRY_KEY}&sentry_version=7&sentry_client=sentry.javascript.node%2F8.26.0"

echo "To test, uncomment the curl command above and update PROJECT_ID and SENTRY_KEY"