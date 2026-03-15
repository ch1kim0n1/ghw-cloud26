# Phase 4 Validation Notes

This folder stores execution notes and evidence for Phase 4 completion.

Primary docs:

- Canonical runbook: `PHASE4_DEMO_RUNBOOK.md`
- Evidence checklist: `phase4-validation/notes/EVIDENCE_TEMPLATE.md`

## Baseline Assets

Use Example1 for the mandatory real-provider baseline run:

- `phase4-validation/input/Example1/video/phase4_test_60s.mp4`
- `phase4-validation/input/Example1/product/product.jpg`
- `phase4-validation/input/Example1/product/metadata.json`

## Evidence Storage Convention

Store run artifacts in a timestamped folder:

- `phase4-validation/notes/evidence-<YYYYMMDD-HHMM>/job.json`
- `phase4-validation/notes/evidence-<YYYYMMDD-HHMM>/preview.json`
- `phase4-validation/notes/evidence-<YYYYMMDD-HHMM>/logs.json`
- `phase4-validation/notes/evidence-<YYYYMMDD-HHMM>/ffprobe-preview.json`

Keep large media outputs in:

- `phase4-validation/output/Example1/`
