# Vultr Provider Notes

The `vultr` provider profile is an explicitly selected backend runtime profile for Phases 2-4.

## Required Services

- analysis service at `VULTR_ANALYSIS_URL`
- LLM completion service at `VULTR_LLM_URL`
- generation service at `VULTR_GENERATION_URL`
- render service at `VULTR_RENDER_URL`
- S3-compatible object storage configured through the `VULTR_OBJECT_STORAGE_*` variables

## HTTP Contracts

### Analysis

- `POST /analysis/jobs`
  - multipart upload with `file`
  - optional fields: `job_id`, `campaign_id`, `product_id`
  - response: `{ "request_id": "analysis_123" }`
- `GET /analysis/jobs/{request_id}`
  - response fields:
    - `request_id`
    - `status`
    - `scenes`
    - optional `payload_ref`
    - optional `message`

### LLM

- `POST /llm/completions`
  - request body mirrors the backend `OpenAIRequest` shape
  - response: `{ "request_id": "llm_123", "content": "{...json...}" }`

### Generation

- `POST /generations`
  - request body mirrors the backend `GenerationRequest` shape
- `GET /generations/{request_id}`
  - response fields must match the backend `GenerationResponse` shape

### Render

- `POST /renders`
  - request body mirrors the backend `RenderRequest` shape
- `GET /renders/{request_id}`
  - response fields must match the backend `RenderResponse` shape

## Artifact Rules

- uploaded artifacts are stored as `s3://bucket/key` references in backend metadata
- the backend remains responsible for copying the final preview back to local storage under `tmp/previews`
- the provider profile must not change job states, slot states, or preview API semantics
