# Notion MCP Audit Integration

## Purpose

This integration mirrors CAFAI pipeline activity into Notion so operators can review full execution history in one workspace.

## What Is Logged

Job-level events are sent to Notion in real time:

- worker transitions across analysis, generation, and render stages
- operator actions: start analysis, select slot, reject slot, re-pick, start generation, start render
- manual generation import events
- terminal failures and error codes when they occur

## Databases Required

Create two databases in Notion and share them with your integration token.

### Jobs database

Required properties:

- Name (title)
- Job ID (rich text)
- Campaign ID (rich text)
- Status (rich text)
- Current Stage (rich text)
- Last Event (rich text)
- Error Code (rich text)
- Updated At (date)
- Summary (rich text)

### Events database

Required properties:

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

## Environment Variables

Backend:

- NOTION_API_BASE_URL (default: https://api.notion.com/v1)
- NOTION_API_KEY
- NOTION_API_VERSION (default: 2022-06-28)
- NOTION_JOBS_DATABASE_ID
- NOTION_EVENTS_DATABASE_ID
- NOTION_REQUEST_TIMEOUT (default: 5s)

Frontend (optional):

- VITE_NOTION_DASHBOARD_URL

## Startup Behavior

- if Notion variables are not set, the audit sink is disabled and the app continues normally
- if Notion variables are set, startup performs connectivity checks to both databases
- startup exits with an error if connectivity validation fails

## Runtime Reliability

- audit writes are asynchronous and buffered
- each write is retried before being marked failed
- audit write failures are logged and do not fail media processing jobs

## Health Endpoint

`GET /api/health` returns audit status:

- enabled: whether audit logging is active
- status: healthy, degraded, or disabled
- details: short diagnostic message

## Demo Runbook

1. Set backend Notion env vars.
2. Set VITE_NOTION_DASHBOARD_URL for convenient UI navigation.
3. Start backend and frontend.
4. Open a job page and click View in Notion.
5. Run analysis, slot selection, generation, and preview render.
6. Confirm Jobs and Events records update in real time.
