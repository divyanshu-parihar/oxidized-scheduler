#!/bin/bash

API_URL="http://localhost:8080/events"

echo "Seeding database with tasks..."

# Helper to send task
send_task() {
    local TYPE=$1
    local TIME=$2
    local PAYLOAD=$3
    local MAX_ATTEMPTS=$4

    # Build JSON using a heredoc to ensure clean formatting and quoting
    JSON_DATA=$(cat <<EOF
{
  "task_type": "$TYPE",
  "scheduled_time": "$TIME",
  "payload": $PAYLOAD,
  "max_attempts": $MAX_ATTEMPTS
}
EOF
)

    curl -X POST "$API_URL" \
      -H "Content-Type: application/json" \
      -d "$JSON_DATA"
    echo -e "\n"
}

# 1. Delayed Notification (+10 minutes)
send_task "send_notification" "$(date -v+10m -u +"%Y-%m-%dT%H:%M:%SZ")" '{"user_id": "u1", "message": "Your trial expires in 3 days"}' 3

# 2. Webhook Dispatcher (Now)
send_task "webhook_dispatch" "$(date -u +"%Y-%m-%dT%H:%M:%SZ")" '{"url": "https://api.partner.com/callback", "event": "order.created", "data": {"id": "ord_99"}}' 5

# 3. Resource Intensive Report Generation (+1 hour)
send_task "generate_report" "$(date -v+1H -u +"%Y-%m-%dT%H:%M:%SZ")" '{"report_type": "monthly_billing", "org_id": "org_55"}' 2

# 4. System Cleanup (+1 day)
send_task "system_cleanup" "$(date -v+1d -u +"%Y-%m-%dT%H:%M:%SZ")" '{"target": "temp_files", "older_than": "7d"}' 1

# 5. Background Data Sync (+5 minutes)
send_task "data_sync" "$(date -v+5m -u +"%Y-%m-%dT%H:%M:%SZ")" '{"source": "hubspot", "destination": "internal_db"}' 10

echo "Seed complete."
