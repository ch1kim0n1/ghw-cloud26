# Figma Software Page Brief

## Purpose

This brief defines a single polished software page for CAFAI using the product and technical docs as the source of truth. It is intended to be translated into a Figma frame or landing-page mockup.

The page should communicate:

- CAFAI is a cloud-assisted contextual ad insertion system
- the workflow is operator-driven and demoable end to end
- the value is seamless in-scene insertion rather than hard-cut ad breaks
- Azure-backed processing is a core part of the product story

## Source References

- `01_Product_Design_Document.md`
- `03_Technical_Specifications.md`
- `06_API_Contracts.md`
- `07_Data_Schema_Definitions.md`
- current frontend shell in `frontend/src/App.tsx`

## Page Type

Single-page marketing and product overview for the software, not the internal dashboard itself.

## Frame Setup

Create two primary frames:

- Desktop: `1440 x 3200`
- Mobile: `390 x 3600`

Use a 12-column desktop grid with `80px` outer margins and `24px` gutters.
Use a 4-column mobile grid with `20px` outer margins and `16px` gutters.

## Visual Direction

The design should feel cinematic, technical, and product-led.

- background: deep navy to slate gradient with subtle film-grain texture
- accents: electric cyan, muted teal, and warm amber for status emphasis
- surfaces: translucent dark cards with soft borders
- shape language: rounded rectangles, restrained glow, thin timeline lines
- mood: premium streaming tooling, not generic SaaS

## Color System

- `bg_primary`: `#08111C`
- `bg_secondary`: `#102235`
- `surface`: `rgba(255,255,255,0.06)`
- `surface_border`: `rgba(255,255,255,0.12)`
- `text_primary`: `#F3F7FB`
- `text_secondary`: `#A7B6C6`
- `accent_cyan`: `#54D1DB`
- `accent_teal`: `#1F8A8A`
- `accent_amber`: `#F4B860`
- `success`: `#5DD39E`

## Type System

Use a sharper editorial heading face paired with a clean product sans.

- display: `Space Grotesk` or `Sora`
- body: `Manrope` or `Plus Jakarta Sans`

Suggested sizes:

- hero display: `64/68`
- section title: `32/38`
- card title: `20/26`
- body: `16/26`
- eyebrow: `12/16` uppercase with `0.12em` tracking

## Page Structure

### 1. Hero

Left side:

- eyebrow: `Context-Aware Fused Ad Insertion`
- headline: `Insert ads that feel like part of the scene.`
- supporting copy: explain that CAFAI analyzes an H.264 MP4, proposes valid insertion slots, generates a short bridge clip, and exports one preview MP4
- primary CTA: `View Workflow`
- secondary CTA: `See MVP Contract`

Right side:

- large product visualization card
- stylized video timeline with three proposed insertion slots
- floating status chips for `Analyzing`, `Generating`, `Stitching`
- a compact panel showing `top 3 slot proposals`

### 2. Problem / Value Strip

Three horizontal value cards:

- `Hard-cut ads break narrative flow`
- `CAFAI finds low-disruption moments inside scenes`
- `Cloud generation creates a context-aware bridge clip`

### 3. Workflow Section

Title:

- `From source video to downloadable preview`

Render a 6-step visual pipeline:

1. Create or select product
2. Upload source video
3. Start analysis explicitly
4. Review ranked slots
5. Approve or edit product line
6. Generate and download preview

Use a left-to-right pipeline on desktop and stacked cards on mobile.

### 4. Operator Dashboard Preview

This section should visually echo the actual app shell.

Include a dashboard mock with:

- top nav: `Products`, `Create Campaign`, `Job`, `Preview`
- health panel
- job progress panel
- slot proposal cards
- product line review panel
- preview panel

This is a presentation mock, not a literal screenshot.

### 5. Why It Works

Two-column section.

Left column:

- short copy on scene-aware placement
- emphasize anchor-frame insertion, not traditional ad breaks

Right column:

- stacked metric cards:
  - `Top 3 ranked slots`
  - `5-8 second bridge clip`
  - `1 downloadable preview MP4`
  - `Queued -> Analyzing -> Generating -> Stitching -> Completed`

### 6. Azure Compute Story

Title:

- `Cloud compute does the heavy work`

Four cards:

- Analysis: `Azure Video Indexer + Azure OpenAI`
- Generation: `Azure ML + Azure OpenAI`
- Audio: `Azure AI Speech`
- Render: `Azure Container Apps + Blob Storage`

Below the cards, add an artifact flow ribbon:

`Generated clip -> Blob temporary storage -> Render worker -> Final preview -> Local download`

### 7. MVP Constraints

Render as a compact spec block with chips:

- `Input: H.264 MP4`
- `Duration: 10-20 min target`
- `Slots: up to 3`
- `Re-picks: up to 2`
- `Preview output: 1 MP4`
- `No auth in MVP`

### 8. Footer CTA

- headline: `Build the first believable in-scene ad workflow`
- support copy: keep it short and technical
- CTA buttons: `Open Docs` and `Review API`

## Component Inventory

Design these reusable components in Figma first:

- primary button
- secondary ghost button
- eyebrow label
- feature card
- metric chip
- timeline slot marker
- status pill
- dashboard panel
- workflow step card
- cloud service card

## Motion Notes

If prototyping in Figma, keep motion restrained:

- hero cards fade and rise in
- slot markers pulse softly
- workflow connector line reveals left to right
- dashboard panels stagger by `80ms`

## Content Notes

Avoid generic AI marketing copy. The language should stay grounded in operator workflow and engineering facts from the docs.

Preferred phrases:

- `context-aware bridge clip`
- `anchor-frame insertion`
- `ranked slot proposals`
- `operator review`
- `downloadable preview MP4`

Avoid:

- `revolutionary`
- `magic`
- `fully autonomous`

## Figma Build Order

1. Create desktop frame and grid
2. Build color and text styles
3. Build reusable cards, chips, and buttons
4. Assemble hero
5. Add workflow and dashboard preview sections
6. Add Azure compute section and MVP spec strip
7. Create mobile adaptation

## Output Expectation

The resulting Figma page should look like a premium product page for a cinematic video AI system, while staying consistent with the repo docs and current dashboard structure.
