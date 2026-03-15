# CAFAI Phase 5: Dynamic Website Ads - Documentation Index

## Overview

This folder contains comprehensive specifications for Phase 5 of the CAFAI platform: **Dynamic Website Ads**. This feature extends CAFAI to generate context-aware AI-generated banner assets for websites, complementing the existing video ad insertion capabilities (Phases 0-4).

## Current Implementation Snapshot

The codebase now includes a shipped subset of this feature:

- frontend upload toggle for `Website Ad`
- synchronous backend generation of banner + vertical outputs
- SQLite persistence for website ad jobs
- local asset storage under `tmp/website_ads`
- frontend showcase pages and injected placement previews

The document set in this folder still represents the broader Phase 5 design target.

If you want the implementation guide for what exists in the repo today, read:

- [`backend/docs/WEBSITE_ADS_FEATURE.md`](../../backend/docs/WEBSITE_ADS_FEATURE.md)

## Reading Order

Recommended reading order for new team members:

1. **[01_Product_Design_Document.md](01_Product_Design_Document.md)** — Read first
   - Product vision and MVP strategy
   - Problem definition and user value proposition
   - Core feature behavior and constraints
   - Success criteria

2. **[02_System_Architecture_Document.md](02_System_Architecture_Document.md)** — Read second
   - High-level system design
   - Control plane architecture
   - Integration with existing CAFAI infrastructure
   - Provider service choices
   - Data flow and processing pipeline

3. **[03_Technical_Specifications.md](03_Technical_Specifications.md)** — Read third
   - Implementation details for backend and frontend
   - API contract specifics
   - Database schema additions
   - Media processing constraints
   - Error handling and retry logic

4. **[04_API_Contracts.md](04_API_Contracts.md)** — Reference
   - Full REST API specification for website ads
   - Request/response formats
   - Error codes and messages

5. **[05_Data_Schema_Definitions.md](05_Data_Schema_Definitions.md)** — Reference
   - SQLite schema extensions
   - Table relationships
   - Indexing strategy

6. **[06_Coding_Standards.md](06_Coding_Standards.md)** — Reference
   - Code organization for Phase 5
   - Naming conventions
   - Testing requirements

7. **[07_Task_Decomposition_Plan.md](07_Task_Decomposition_Plan.md)** — Reference
   - Ordered Phase 5 build plan
   - Deliverables per sprint
   - Exit criteria

## Document Purposes

| Document | Purpose | Audience |
|----------|---------|----------|
| 01_Product_Design_Document.md | Define what the product does and why | PMs, designers, engineers |
| 02_System_Architecture_Document.md | Define how the system is structured | Architects, senior engineers |
| 03_Technical_Specifications.md | Define implementation details | Engineers |
| 04_API_Contracts.md | Define the HTTP API surface | Frontend engineers, integrators |
| 05_Data_Schema_Definitions.md | Define database schema | Backend engineers, DBAs |
| 06_Coding_Standards.md | Define code organization and patterns | All engineers |
| 07_Task_Decomposition_Plan.md | Define the build plan and sprints | Tech leads, project managers |

## Key Assumptions

**All content in this folder assumes that Phases 0-4 (CAFAI Video Ad Insertion) are fully complete and operational.**

- Existing CAFAI infrastructure is available
- Video ad insertion pipeline is production-ready
- Backend, frontend, and database layers are stable
- Cloud provider integrations (Azure + Vultr) are proven and working

## Quick Start for Developers

**New to Phase 5?**

1. Read the shipped implementation guide in `backend/docs/WEBSITE_ADS_FEATURE.md`
2. Start with 01_Product_Design_Document.md
3. Skim 02_System_Architecture_Document.md
4. Read 03_Technical_Specifications.md in full
5. Use 04_API_Contracts.md and 05_Data_Schema_Definitions.md as reference

**Total onboarding time: ~1 hour for the design set, plus ~10 minutes for the shipped implementation guide**

## Status

- **Phase 5 Status:** Partially implemented, with broader design work still documented here
- **Last Updated:** March 2026
- **Owner:** CAFAI Platform Team

---

For questions or clarifications, refer to the relevant specification document or consult with the platform team.
