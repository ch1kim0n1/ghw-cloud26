# Product Design Document

## Product

Real-Time Dynamic Movie Ad Engine

## MVP Strategy

Context-Aware Fused Ad Insertion (CAFAI)

## 1. Canonical MVP Statement

The MVP analyzes a provided 10-20 minute H.264 MP4, proposes up to the top 3 candidate anchor-frame insertion slots inside scenes, lets the operator review those slots, optionally edit the generated product line, generates one 5-8 second CAFAI bridge clip in which the on-screen character naturally interacts with the advertised product, inserts that clip between the chosen anchor frames with basic audio continuity, and exports one downloadable preview MP4.

## 2. Product Goal

Build a cloud-assisted system that can place a context-aware ad moment inside existing video in a way that feels like part of the scene instead of a traditional ad break.

The MVP must prove:

- the system can analyze a full input clip end to end
- the system can pick believable insertion moments automatically
- the generated ad moment can visually and narratively fit the surrounding scene
- the inserted segment can be stitched into the source video with coherent enough audio to feel continuous
- the cloud pipeline is necessary for the heavy analysis, generation, audio, and rendering work

## 3. Problem Definition

Traditional ads interrupt narrative flow. They appear at arbitrary breakpoints, ignore scene context, and feel separate from the content the viewer is already watching.

CAFAI targets a different behavior:

- inspect the actual video content
- find low-disruption moments inside scenes
- generate a short ad moment that uses the surrounding narrative context
- continue the original movie after the inserted moment with minimal disruption

The intended viewer perception is not "the player paused for an ad." The intended perception is "this moment felt like part of the movie."

## 4. Target Users

### Primary Users

- streaming platform product teams
- ad-tech teams exploring premium contextual placements
- media technology teams building monetization tooling
- post-production and content operations teams

### Secondary Users

- advertisers testing immersive ad formats
- research and demo teams working on AI video augmentation
- hackathon judges, technical evaluators, and investors

## 5. User Value Proposition

The MVP should provide:

- less disruptive ad experiences than hard-cut ad breaks
- higher-value placements because the ad appears inside scene context
- a new monetization approach for existing content libraries
- a clear demo of why cloud compute is needed for analysis, generation, audio, and rendering

## 6. Core MVP Behavior

### 6.1 Input

- one source video: H.264 MP4
- target production scope: 10-20 minutes
- one advertised product, either pre-created in the product catalog or created during campaign setup

### 6.2 Analysis

The system does not start analysis automatically on campaign creation.

The operator:

- creates a campaign
- uploads a source video
- explicitly starts analysis

The system then:

- analyzes the full uploaded video
- segments scenes
- identifies candidate anchor-frame pairs inside scenes
- excludes bad insertion regions
- ranks up to the top 3 candidate slots

### 6.3 Slot Review

The operator can:

- inspect the returned slot proposals
- select one slot
- reject one or more slots
- request a re-pick up to 2 times if all proposed slots are rejected

If analysis finds fewer than 3 valid slots:

- return the available valid slots
- fail only when no valid slots are found

### 6.4 Product Line Review

After the operator selects a slot, the system generates a suggested product mention line from product metadata and scene context.

The operator can then:

- accept the generated line
- edit the line
- disable dialogue and request a silent interaction instead

### 6.5 Generation

For the selected slot, after product line review is complete, the system generates a CAFAI clip that:

- begins from the chosen start anchor frame
- resolves back into the chosen finish anchor frame
- shows the product directly in the generated clip
- includes either a short spoken mention or a silent product interaction depending on what better fits the scene
- preserves the surrounding narrative context as much as possible

### 6.6 Stitching

The system inserts extra duration into the source video:

- the original runtime increases by the inserted clip duration
- the source frames are not replaced
- the inserted clip sits between the two anchor frames
- audio before, during, and after the inserted clip must remain basically coherent

### 6.7 Storage

Azure Blob Storage is used as temporary cloud artifact storage during generation and rendering. After rendering completes, the final preview file is copied back to local storage for MVP download, inspection, and debugging.

### 6.8 Operator Workflow

The dashboard must allow the operator to:

- create a product or choose an existing one
- create a campaign and upload a video
- explicitly start analysis
- monitor job progress
- inspect the proposed slots
- select, reject, and re-pick slots
- review or edit the suggested product line
- watch generation progress
- watch render progress
- preview and download the final output MP4

## 7. MVP Exclusion Rules

The MVP rejects candidate slots when any of the following is true:

- normalized motion score across the candidate window is greater than `0.65`
- either anchor boundary sub-window exceeds motion score `0.75`
- action intensity score across the candidate window is greater than `0.70`
- any 1-second sub-window exceeds action intensity `0.80`
- a shot boundary is detected within `0.5` seconds of either anchor
- cut-confidence at either anchor exceeds `0.70`
- scene duration is less than `10` seconds
- no quiet window of at least `3` seconds exists

These are MVP heuristic defaults and may be tuned after early test runs.

## 8. Constraints

### Product Constraints

- one insertion per job
- one preview output per job
- default generated ad duration: `6` seconds
- preferred generated ad duration: `5-8` seconds
- maximum generated ad duration: `8` seconds
- no invisible auto-publish behavior; the operator explicitly starts analysis and confirms the slot workflow

### Technical Constraints

- local storage is used for uploaded assets and final preview download in MVP
- Azure Blob Storage is used only for temporary cloud artifacts
- SQLite is used for MVP metadata persistence
- no authentication in MVP
- coarse job states only: `queued -> analyzing -> generating -> stitching -> completed|failed`
- generation failure fails the job
- render failure fails the job but preserves generated artifacts for retry

### Cloud Constraints

Cloud compute is used for:

- video analysis
- context and narrative understanding
- top slot proposal generation
- CAFAI clip generation
- audio generation and alignment
- final preview rendering

## 9. Azure Service Choices

The MVP documentation assumes:

- analysis: Azure Video Indexer + Azure OpenAI
- CAFAI generation: Higgsfield Kling + Azure OpenAI as primary, Azure Machine Learning as fallback
- audio generation and alignment: Azure AI Speech
- final render: Azure Container Apps running ffmpeg + Azure Blob Storage for temporary artifacts

## 10. Non-Goals

The MVP does not attempt to solve:

- production auth and access control
- multi-tenant architecture
- live streaming playback integration
- runtime-neutral insertion
- advanced photorealistic quality guarantees
- legal or licensing workflows
- personalized ad targeting
- ad auctions or demand-side integrations

Post-MVP work must begin only after the MVP is stable and demoable.

## 11. Success Metrics

### Primary MVP Success

- the inserted ad moment feels seamless enough that the viewer does not immediately read it as a traditional ad break
- the demo makes the cloud pipeline obvious and necessary

### Operational MVP Success

- one representative input video completes the full flow end to end
- the system proposes valid insertion slots automatically
- the operator can reject and re-pick slots
- the operator can accept, edit, or disable the product line
- one selected slot generates successfully
- the final preview MP4 renders and downloads successfully

## 12. Risks

- generated character interaction may not match the source scene convincingly
- generated speech may feel unnatural or tonally inconsistent
- anchor-frame continuity may still break on difficult scenes
- cloud generation latency or provider quality may impact the demo
- narrative fit is subjective and may require careful clip selection for the hackathon

## 13. Future Extensions

After MVP only:

- fallback generation or composition path
- stronger narrative-sensitivity modeling
- better photorealistic character-product interaction
- multi-ad planning
- streaming platform integration
- authentication, tenancy, and quotas
- production cloud storage and database migration
