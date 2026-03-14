# Figma Software Page Brief

## 1. Purpose

This document is the design and frontend handoff brief for one polished CAFAI software page.

The output is:

- a single-page marketing and product overview
- visually premium and presentation-ready
- grounded in the real MVP behavior in this repository
- detailed enough for a frontend developer or designer to build without guessing core structure, copy intent, or interaction behavior

This page is not the internal operator dashboard itself. It is the product-story page that explains the system clearly and shows the dashboard as one part of the story.

## 2. Page Goal

The page must make a first-time viewer understand all of the following within one scroll:

- what CAFAI is
- why it is different from hard-cut ad breaks
- how the operator workflow works end to end
- why cloud-backed analysis, generation, and rendering are required
- what the MVP actually outputs
- what the real software interface looks like

The page should make the software look ambitious, credible, and demoable now.

## 3. Product Truths To Preserve

All page content must stay consistent with the real repo docs and current implementation.

The page must communicate these truths:

- CAFAI means Context-Aware Fused Ad Insertion
- the system analyzes an uploaded H.264 MP4
- it proposes up to the top 3 ranked insertion slots when possible
- the operator explicitly starts analysis
- the operator can review, reject, and re-pick slots
- the operator can manually enter a slot window after analysis if needed
- the operator can accept, edit, or disable the product line
- the system generates a short context-aware bridge clip
- the generated clip is inserted between anchor frames
- the output is one downloadable preview MP4
- the workflow is cloud-assisted, not purely local
- Azure is still the default product-story provider path for analysis, rendering, and fallback generation

Do not imply:

- autonomous publishing
- final production readiness
- multiple previews per job
- live-stream insertion
- personalized ad targeting

## 4. Intended Audience

Primary audience:

- hackathon judges
- technical evaluators
- investors
- product and ad-tech stakeholders

Secondary audience:

- frontend and design contributors who need a precise visual/content target

The visual and copy tone should assume the viewer is technical, visually literate, and impatient.

## 5. Deliverable Type

Create one single-page software overview.

The page should feel like:

- a premium product page
- a cinematic technology demo
- a credible software showcase

It should not feel like:

- a generic SaaS landing page
- a docs site
- the actual dashboard application

## 6. Frame Setup

Create two primary frames:

- Desktop: `1440 x 3200`
- Mobile: `390 x 3600`

Grid:

- desktop: 12 columns, `80px` outer margin, `24px` gutters
- mobile: 4 columns, `20px` outer margin, `16px` gutters

Spacing system:

- section spacing desktop: `96px`
- section spacing mobile: `64px`
- card padding desktop: `24px`
- card padding mobile: `18px`
- small gap: `8px`
- medium gap: `16px`
- large gap: `24px`
- extra large gap: `32px`

## 7. Visual Direction

The page should feel cinematic, technical, and product-led.

Creative direction:

- premium streaming-tooling atmosphere
- visual confidence without looking noisy
- dark, layered, and deliberate
- more studio control room than startup dashboard

Visual principles:

- use depth, gradients, translucency, and restrained glow
- make timelines, slots, and workflow steps visually legible
- keep cards clean and information-dense
- preserve enough air around key headlines and hero content

Avoid:

- flat white layouts
- generic startup illustration style
- excessive neon everywhere
- abstract shapes that do not support the product story

Design-director note:

- the page should feel like a premium film-tech control surface translated into a product page
- it should suggest confidence, precision, and editorial taste
- it should not feel playful, cute, or app-store generic

## 8. Color System

Use this as the default palette.

- `bg_primary`: `#08111C`
- `bg_secondary`: `#102235`
- `bg_tertiary`: `#132B42`
- `surface`: `rgba(255,255,255,0.06)`
- `surface_strong`: `rgba(255,255,255,0.10)`
- `surface_border`: `rgba(255,255,255,0.12)`
- `text_primary`: `#F3F7FB`
- `text_secondary`: `#A7B6C6`
- `text_muted`: `#7D91A5`
- `accent_cyan`: `#54D1DB`
- `accent_teal`: `#1F8A8A`
- `accent_amber`: `#F4B860`
- `accent_red`: `#D96B6B`
- `success`: `#5DD39E`

Usage guidance:

- hero and page background should use a navy-to-slate gradient
- cyan should highlight analysis and system intelligence
- amber should highlight selected moments, timeline emphasis, and “insertion”
- success green should be used sparingly for completed states
- red should appear only for rejection/failure chips in small amounts

## 9. Typography

Use a sharper editorial display face paired with a clean technical sans.

Recommended type pairing:

- display/headline: `Space Grotesk` or `Sora`
- body/UI: `Manrope` or `Plus Jakarta Sans`

Suggested scale:

- hero display: `64/68`
- hero support: `18/30`
- section title: `32/38`
- section intro: `18/30`
- card title: `20/26`
- body: `16/26`
- small body: `14/22`
- eyebrow: `12/16`, uppercase, `0.12em` tracking

Weight guidance:

- hero headline: `700`
- section titles: `700`
- card titles: `600`
- body: `400` to `500`

## 10. Background and Atmosphere

The page background should not be a flat fill.

Required treatment:

- base dark gradient
- one subtle radial highlight behind the hero
- low-contrast film-grain or noise texture
- optional soft horizontal linework or timeline motifs

Optional treatment:

- blurred cyan and amber light blooms behind hero media
- subtle grid or scanline texture in the workflow section

Atmosphere target:

- think "high-end post-production suite" rather than "consumer AI app"
- think "streaming platform prototype for executives" rather than "hackathon flyer"

## 10.1 Material Language

Surfaces should feel like layered glass, smoked acrylic, and premium monitoring interfaces.

Use:

- dark translucent panels
- hairline borders
- restrained internal highlights
- soft shadow depth
- sparse accent glows only where emphasis is needed

Do not use:

- heavy cartoon shadows
- loud gradient borders on every card
- pill buttons that look toy-like
- oversaturated holographic effects

## 10.2 Image and Illustration Language

Do not rely on generic AI art or stock-illustration tropes.

Preferred visual motifs:

- timeline rails
- frame markers
- slot indicators
- playback windows
- subtle waveform or transcript cues
- anchor-point brackets
- orchestration-style system diagrams

If using imagery inside hero or dashboard mock areas:

- keep it stylized and software-centric
- avoid realistic human faces as the primary focal point
- let the software process and insertion logic remain the hero

## 11. Core Story

The page should tell this narrative in order:

1. Traditional ad breaks interrupt scenes.
2. CAFAI finds better moments inside scenes.
3. The operator stays in control of the process.
4. Cloud analysis and generation do the heavy work.
5. The result is one believable inserted moment and one preview MP4.

Every section should reinforce one part of this story.

## 12. Page Structure

Use this exact section order.

### 12.1 Hero

Purpose:

- communicate the value proposition in one screen
- make the software feel premium immediately
- show the product is both cinematic and operational

Desktop layout:

- left column: copy and CTAs
- right column: visual system card

Mobile layout:

- stack copy above visual

Required copy:

- eyebrow: `Context-Aware Fused Ad Insertion`
- headline: `Insert ads that feel like part of the scene.`
- support copy: `CAFAI analyzes an uploaded H.264 MP4, proposes ranked insertion slots, generates a short context-aware bridge clip, and exports one downloadable preview MP4.`
- primary CTA: `View Workflow`
- secondary CTA: `See MVP Contract`

Hero visual content must include:

- a stylized video timeline
- exactly three visible slot markers
- status pills for `Analyzing`, `Generating`, and `Stitching`
- one compact “Top 3 slot proposals” panel
- one visual indication that a bridge clip is inserted between anchors

Hero visual should suggest:

- ranking
- operator review
- controlled insertion

It should not look like:

- an ad player
- a social media editor
- a generic AI image generator

Hero art-direction note:

- the right side should feel like the strongest frame on the page
- this is the page's "poster shot"
- if only one visual is remembered, it should be the hero timeline/insertion composition

Hero composition guidance:

- keep one dominant visual object
- keep supporting panels secondary and layered around it
- use asymmetry so the page does not feel like a standard product grid
- make the slot markers and inserted bridge segment readable in under 2 seconds

### 12.2 Problem / Value Strip

Purpose:

- explain why this exists before the viewer needs to infer it

Render three horizontal cards on desktop and stacked cards on mobile.

Exact card copy:

- `Hard-cut ads break narrative flow`
- `CAFAI finds low-disruption moments inside scenes`
- `Cloud generation creates a context-aware bridge clip`

Each card should include:

- short supporting line, 1 sentence maximum
- one subtle icon or visual cue

Suggested support copy:

- hard-cut card: `Traditional breaks ignore scene context and feel external to the story.`
- slot card: `CAFAI ranks anchor-frame insertion points with continuity and quiet-window heuristics.`
- generation card: `The system produces a short ad moment that resolves back into the scene.`

### 12.3 Workflow Section

Purpose:

- explain the real operator flow end to end
- make the system feel actionable and demoable

Section title:

- `From source video to downloadable preview`

Section intro:

- `The operator stays in control while cloud services handle analysis, generation, and rendering.`

Render a 6-step pipeline.

Exact step labels:

1. `Create or select product`
2. `Upload source video`
3. `Start analysis explicitly`
4. `Review ranked slots`
5. `Approve or edit product line`
6. `Generate and download preview`

Each step card should include:

- step number
- short label
- one-line description

Suggested step descriptions:

1. `Use an existing product or create one from metadata, image, or source URL.`
2. `Upload one H.264 MP4 and attach it to a campaign.`
3. `The workflow begins only when the operator explicitly starts analysis.`
4. `Review top-ranked insertion candidates, reject weak fits, and request a re-pick if needed.`
5. `Accept the suggested line, edit it, or disable dialogue entirely.`
6. `Render one preview MP4 and review the inserted result.`

Desktop behavior:

- show left-to-right progression with connector line

Mobile behavior:

- stacked vertical cards
- connector line becomes vertical

Art-direction note:

- this section should feel crisp and operational, like a systems view
- make the progression feel inevitable and legible, not decorative

### 12.4 Operator Dashboard Preview

Purpose:

- show the real software shape
- bridge the product story and the actual app

This section is a curated product mock of the dashboard, not a screenshot.

Section title:

- `Operator review stays in the loop`

Section intro:

- `The dashboard exposes each major control point: intake, analysis, slot review, line review, and preview output.`

Required dashboard elements:

- top nav with `Products`, `Create Campaign`, `Job`, `Preview`
- health panel
- job progress panel
- slot proposal cards
- product line review panel
- preview render panel

Important:

- show a selected slot clearly
- show at least one rejected slot state
- show one line-review state
- show one preview-ready state

This section should make the software look real, not conceptual.

Art-direction note:

- this should feel like a product reveal, not a wireframe
- use enough density to feel believable
- avoid over-detail that turns the section into unreadable miniature UI

Dashboard mock fidelity rules:

- panels should look like they belong to the same design system
- each panel should show one clear function
- one selected item should be obvious immediately
- one "completed" or "preview ready" state should provide emotional payoff

### 12.5 Why It Works

Purpose:

- explain the system logic in plain product terms

Two-column layout.

Left column:

- headline: `Why the insertion feels like part of the scene`
- short paragraph about scene-aware placement, anchor continuity, and operator review

Recommended copy direction:

- emphasize anchor-frame insertion rather than hard cuts
- emphasize believable context rather than generic ad insertion

Right column:

- four stacked metric cards

Exact metrics:

- `Top 3 ranked slots`
- `5-8 second bridge clip`
- `1 downloadable preview MP4`
- `Queued -> Analyzing -> Generating -> Stitching -> Completed`

Each card should include one short support line.

Art-direction note:

- this section should slow the pacing slightly
- let it feel like the product's thesis, not another feature row

### 12.6 Cloud Compute Story

Purpose:

- make it obvious why the cloud pipeline matters

Section title:

- `Cloud compute does the heavy work`

Section intro:

- `The local control plane stays simple while analysis, generation, audio, and rendering run through cloud-backed services.`

Render four service cards:

- `Analysis` / `Azure Video Indexer + Azure OpenAI`
- `Generation` / `Higgsfield Kling + Azure OpenAI, with Azure ML fallback`
- `Audio` / `Azure AI Speech`
- `Render` / `Azure Container Apps + Blob Storage`

Below the cards, render the artifact flow ribbon exactly as:

`Generated clip -> Blob temporary storage -> Render worker -> Final preview -> Local download`

This ribbon should be very legible and should visually connect the services to the final preview.

Art-direction note:

- this section should feel engineered, not abstract
- the viewer should leave understanding where the heavy compute happens

### 12.7 MVP Constraints

Purpose:

- show the product as credible and scoped, not vague

Title:

- `MVP Scope`

Render as compact chips or short spec cards.

Exact items:

- `Input: H.264 MP4`
- `Duration: 10-20 min target`
- `Slots: up to 3`
- `Re-picks: up to 2`
- `Preview output: 1 MP4`
- `No auth in MVP`

Optional note below:

- `A smaller 40-60 second baseline validation profile is also used for engineering validation before full-length runs.`

### 12.8 Footer CTA

Purpose:

- close the page cleanly and push the viewer toward docs or review

Required copy:

- headline: `Build the first believable in-scene ad workflow`
- support copy: `CAFAI combines operator control, ranked scene analysis, and cloud-assisted generation into one demoable insertion pipeline.`
- primary CTA: `Open Docs`
- secondary CTA: `Review API`

## 13. Required Content Coverage Checklist

The finished page must visibly communicate all of these:

- one uploaded source video
- ranked slot proposals
- operator control
- reject and re-pick behavior
- product line review
- generated bridge clip
- anchor-frame insertion
- preview render and download
- cloud-backed analysis/generation/rendering
- one preview MP4 output

If one of these is not visible or clearly stated, the page is incomplete.

## 13.1 Priority Hierarchy

If visual tradeoffs are required, preserve this hierarchy in order:

1. value proposition
2. workflow clarity
3. dashboard credibility
4. cloud-compute explanation
5. MVP scope details

Do not sacrifice the first three for decorative flourish.

## 14. Component Inventory

Create these reusable components first.

- primary button
- secondary ghost button
- eyebrow label
- feature card
- workflow step card
- metric chip
- metric card
- status pill
- timeline slot marker
- timeline bridge segment
- dashboard panel
- cloud service card
- spec chip
- CTA block

## 15. Component State Rules

Even though this is a marketing page, component states must be defined so the frontend build is consistent.

Buttons:

- default
- hover
- pressed
- focus-visible
- disabled

Status pills:

- analyzing
- generating
- stitching
- completed
- rejected

Cards:

- default
- hover-lift optional on desktop
- no hover-only dependency on mobile

State styling should remain subtle and premium.

- hover should feel like a controlled lift, not a bounce
- focus should feel intentional and visible
- selected states should be more about contrast and border emphasis than loud effects

## 16. Responsive Rules

Desktop:

- hero uses two columns
- workflow uses horizontal pipeline
- dashboard preview may use multi-panel composition
- cloud section uses four cards in a row or two-by-two if needed

Tablet:

- reduce hero visual density before collapsing to single column
- workflow may become two-row grid before full stack

Mobile:

- all sections stack vertically
- hero CTAs stack or wrap
- timeline visual simplifies but still shows 3 slot markers
- dashboard preview becomes one stacked mock composition
- artifact flow ribbon becomes vertical or segmented chips

No section should rely on tiny text or micro-detail to remain understandable on mobile.

## 17. Motion Guidance

If motion is prototyped or implemented, keep it restrained.

Allowed motion:

- hero cards fade and rise in
- slot markers pulse softly
- workflow connector reveals left to right
- dashboard panels stagger by `80ms`
- status pills can shimmer subtly

Motion constraints:

- no looping motion that becomes distracting
- no parallax dependency
- support reduced-motion behavior by disabling non-essential animation

Motion should feel like interface confidence, not marketing theater.

## 18. Accessibility Requirements

The design must be buildable accessibly.

Required:

- strong enough contrast on all text and chips
- clear heading hierarchy
- buttons and links must remain obvious without color alone
- decorative visuals must not carry core meaning alone
- animation must remain optional
- CTA labels must be explicit

The page should still make sense if the user only reads the headings and short body copy.

## 19. Copy Rules

Tone:

- technical
- concise
- confident
- grounded

Preferred phrases:

- `context-aware bridge clip`
- `anchor-frame insertion`
- `ranked slot proposals`
- `operator review`
- `downloadable preview MP4`
- `cloud-assisted`

Avoid:

- `revolutionary`
- `magic`
- `fully autonomous`
- `disruptive`
- vague AI hype language

Voice reference:

- write like a strong creative director partnered with a technical product lead
- every headline should feel sharp enough for a keynote slide
- every support line should still survive engineering review

## 20. Frontend Implementation Notes

This document is intentionally specific enough to code from.

A frontend developer should be able to map each section directly into:

- one page container
- one section component
- reusable cards, pills, chips, and CTA components
- one dark theme token set
- responsive layout rules for desktop and mobile

If engineering needs implementation-level structure, the page can be split into:

- `HeroSection`
- `ValueStripSection`
- `WorkflowSection`
- `DashboardPreviewSection`
- `WhyItWorksSection`
- `CloudComputeSection`
- `MvpScopeSection`
- `FooterCtaSection`

## 21. What “Best Version Of The Software” Means Here

The page should showcase CAFAI at its best by making these things feel strong and clear:

- the insertion is scene-aware
- the operator remains in control
- the workflow is end to end
- the cloud stack is necessary and visible
- the dashboard looks real and usable
- the output is concrete: one preview MP4

The page should not exaggerate quality claims that the repo does not support. It should present the MVP in its strongest truthful form.

Specifically, the page should emphasize:

- believable insertion logic over raw AI spectacle
- product control over automation fantasy
- one polished end-to-end path over a scattered feature set
- operator trust over novelty

The page should make CAFAI look like:

- a serious new ad format prototype
- a strong hackathon finalist product
- a credible internal concept ready for deeper investment

The page should not make CAFAI look like:

- a generic video editor
- a prompt-to-video toy
- an unfinished research demo with no operator workflow

## 21.1 Final Visual Test

Before approving the design, ask:

- Does the hero feel memorable?
- Does the workflow read in one pass?
- Does the dashboard mock feel real?
- Does the page explain why the cloud stack matters?
- Would a judge understand the product without narration?
- Would a frontend developer know what to build without inventing the visual story?

If the answer to any of these is no, the design is not finished.

## 22. Final Acceptance Criteria

This brief is satisfied when the resulting page:

- looks premium and intentional on desktop and mobile
- communicates the software in one pass without extra explanation
- includes every required section in the specified order
- accurately reflects the MVP behavior in the repo docs
- gives a frontend developer enough clarity to implement without inventing the product story
- makes CAFAI feel like a real, compelling software system rather than a concept
