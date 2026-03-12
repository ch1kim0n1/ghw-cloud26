# Coding Standards Document

## 1. Purpose
This document defines coding and architecture rules for the project so implementation remains consistent across backend, frontend, workers, and infrastructure code.

## 2. General Principles
- prefer clarity over cleverness
- keep modules small and bounded
- do not mix orchestration logic with core business logic
- avoid hidden side effects
- all public interfaces must be typed and validated
- all failures must produce structured errors

## 3. Naming Conventions
### Python
- files/modules: `snake_case`
- functions: `snake_case`
- variables: `snake_case`
- classes: `PascalCase`
- constants: `UPPER_SNAKE_CASE`

### TypeScript
- files for components: `PascalCase.tsx`
- files for utilities/hooks: `camelCase.ts` or `useThing.ts`
- variables/functions: `camelCase`
- classes/types/interfaces: `PascalCase`
- constants: `UPPER_SNAKE_CASE`

### API / Schema Fields
- JSON keys: `snake_case` or `camelCase`, but only one standard per API version
- choose one convention and enforce it globally

## 4. Architecture Patterns
- controllers should be thin
- services should contain business logic
- repositories should contain persistence access only
- workers should orchestrate async stage execution, not hide domain rules
- adapters should isolate vendor/cloud SDK specifics
- shared contracts must be source-of-truth for interfaces

## 5. Function and Class Design
- functions should do one thing
- functions longer than ~50 lines should be reviewed for extraction
- classes must have explicit responsibility
- prefer composition over inheritance unless inheritance is clearly justified
- avoid God services

## 6. Error Handling
### Rules
- never swallow exceptions silently
- wrap external dependency failures with project-specific context
- return structured error objects from APIs
- distinguish between user errors, system errors, and transient infra errors

### Structured Error Format
```json
{
  "error": {
    "code": "RENDER_FAILED",
    "message": "Render stage failed",
    "details": {
      "job_id": "...",
      "stage": "rendering"
    }
  }
}
```

## 7. Logging Standards
### Rules
- use structured logs only
- include correlation identifiers in every service boundary log
- never log secrets or raw credentials
- log state transitions explicitly

### Required Context Fields
- `job_id`
- `request_id`
- `service`
- `stage`
- `status`
- `duration_ms` where relevant

## 8. Formatting and Linting
### Python
- formatter: `black`
- import sorting: `isort`
- linting: `ruff` or `flake8`
- typing: `mypy` where practical

### TypeScript
- formatter: `prettier`
- linting: `eslint`
- strict typing enabled

## 9. Comments and Documentation
- comment why, not what, unless the code is non-obvious
- public functions/classes should have concise docstrings where useful
- every service package should include a README if its behavior is non-trivial
- TODO comments must include owner or issue reference where possible

## 10. API Standards
- APIs must be versioned
- every request payload must be schema-validated
- every response must follow explicit contract definitions
- timestamps must use ISO-8601 UTC unless a different format is explicitly required
- pagination and filtering should use consistent query conventions

## 11. Testing Standards
### Minimum Expectations
- unit tests for core domain logic
- integration tests for repositories and service handoffs
- contract tests for API/event compatibility
- e2e tests for critical flows

### Rules
- tests must be deterministic
- avoid network dependence in unit tests
- mock third-party services at boundaries
- use fixtures for representative video/job metadata

## 12. Configuration Standards
- config must come from environment or config files, never hardcoded secrets
- feature flags must be explicit
- environment-specific behavior must be isolated in config layers

## 13. Git and Review Standards
- one concern per PR when possible
- PRs must include test evidence
- breaking contract changes require explicit review callout
- generated files should not be committed unless intentionally versioned

## 14. Security Standards
- no secrets in repository
- signed URL or token-based asset access only
- validate uploaded file types and size limits
- sanitize user-provided metadata before downstream use

## 15. Cloud / Worker Standards
- background jobs must be idempotent where possible
- retryable failures must be marked explicitly
- GPU jobs must emit usage and timing metrics
- all object storage writes must validate success before committing DB state
