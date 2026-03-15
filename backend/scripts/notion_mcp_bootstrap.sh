#!/usr/bin/env bash
set -euo pipefail

cat <<'EOF'
Notion MCP bootstrap checklist for CAFAI audit logging

1) Create two Notion databases in your workspace:
   - Jobs database (example name: CAFAI Jobs)
   - Events database (example name: CAFAI Events)

2) Required Jobs database properties:
   - Name (title)
   - Job ID (rich text)
   - Campaign ID (rich text)
   - Status (rich text)
   - Current Stage (rich text)
   - Last Event (rich text)
   - Error Code (rich text)
   - Updated At (date)
   - Summary (rich text)

3) Required Events database properties:
   - Event (title)
   - Job ID (rich text)
   - Campaign ID (rich text)
   - Event Type (rich text)
   - Status (rich text)
   - Current Stage (rich text)
   - Error Code (rich text)
   - Timestamp (date)
   - Message (rich text)
   - Metadata (rich text)

4) Set environment variables:
   export NOTION_API_KEY="secret_..."
   export NOTION_API_BASE_URL="https://api.notion.com/v1"
   export NOTION_API_VERSION="2022-06-28"
   export NOTION_JOBS_DATABASE_ID="..."
   export NOTION_EVENTS_DATABASE_ID="..."
   export NOTION_REQUEST_TIMEOUT="5s"

5) Share both databases with your integration token.

6) Start backend and create a job. You should see real-time records in Notion.
EOF

missing=0
for name in NOTION_API_KEY NOTION_JOBS_DATABASE_ID NOTION_EVENTS_DATABASE_ID; do
  if [[ -z "${!name:-}" ]]; then
    printf 'WARN: %s is not set\n' "$name"
    missing=1
  fi
done

if [[ $missing -eq 0 ]]; then
  echo "Notion audit environment variables look configured."
else
  echo "Set the missing variables before running backend for Notion logging."
fi
