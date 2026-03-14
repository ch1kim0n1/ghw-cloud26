# Phase 5: Dynamic Website Ads - Product Design Document

## Product

Context-Aware AI-Generated Website Ad Assets

## Extension Strategy

Phase 5 extends CAFAI from video ad insertion to dynamic website ad generation, leveraging the existing content analysis and AI generation capabilities to create contextual banner assets.

## 1. Canonical Phase 5 Statement

The Phase 5 MVP accepts one article URL or article text, accepts one advertised product with metadata and optional image, analyzes the article content for themes and context, generates up to 3 candidate AI-rendered ad banner designs that contextually blend article themes with the product being advertised, lets the operator review and select one design, generates multiple asset variants (square banner, vertical banner, icon), and exports downloadable PNG assets ready for web deployment.

## 2. Product Goal

Extend CAFAI's core competency (context-aware AI generation) from video to the web advertising domain, enabling marketers and content platforms to quickly generate contextually relevant, on-brand ad banners that feel thematically connected to surrounding editorial content instead of generic promotional graphics.

The Phase 5 MVP must prove:

- the system can analyze article content and extract semantic themes
- the system can map product characteristics to article context
- the system can generate visually coherent banner designs that blend both contexts
- the generated banners can be rendered in multiple formats (square, vertical, icon)
- the web ad asset pipeline reuses existing CAFAI orchestration patterns

## 3. Problem Definition

Current state of web advertising:

- **Generic banners:** Most website ads are static images that don't adapt to surrounding content
- **Context mismatch:** Ads appear randomly, often disconnected from article themes
- **High production cost:** Creating custom ad variations for different contexts is expensive and time-consuming
- **Low engagement:** Generic ads blend into page clutter and generate poor click-through rates

Phase 5 targets a different approach:

- analyze the actual article content
- extract semantic themes (e.g., "historic Rome," "classical architecture," "ancient history")
- generate brand-aware ad designs that blend product with article context
- render multiple formats optimized for different placements
- reduce time-to-deployment from weeks to minutes

Example context blending:

```
Article: "The Colosseum: Engineering Marvel of Ancient Rome"
Product: Premium wireless headphones
Generated Banner: Julius Caesar in a toga, listening to music with the advertised headphones,
                  Roman architecture in background
Context: Historic + technology = Creates narrative connection
```

## 4. Target Users

### Primary Users

- **Marketing teams** at ad-tech platforms and programmatic networks
- **Content platforms** (news sites, blogs, long-form publishers) seeking new revenue streams
- **E-commerce retailers** wanting contextual ad campaigns across partner content
- **Creative agencies** building customized ad campaigns at scale

### Secondary Users

- **In-house marketing teams** at mid-size publishers
- **Ad networks** building marketplace inventory
- **Research and demo teams** evaluating AI-assisted creative workflows

## 5. User Value Proposition

Phase 5 should provide:

- **Faster asset creation:** Generate on-brand ad banners in minutes instead of weeks
- **Higher relevance:** Ads that thematically match surrounding content
- **Cost reduction:** AI generation cheaper than hiring designers for custom assets
- **Format flexibility:** One creative concept rendered across square, vertical, and icon formats
- **Differentiation:** Demonstrate multi-format capability beyond video CAFAI

## 6. Core Phase 5 Behavior

### 6.1 Input

**Article Source (one of):**
- Direct article URL (system fetches and analyzes)
- Direct article text/markdown (user provides)
- Article headline + body text (user provides)

**Advertised Product:**
- Product name (required)
- Product description (required)
- Product image or URL (optional)
- Category/context keywords (optional)
- Brand voice/tone guidance (optional)

### 6.2 Processing Stages

#### Stage 1: Article Analysis

The system:
- Fetches or ingests article content
- Extracts primary themes (historic period, location, industry, mood)
- Identifies key visual elements (settings, time periods, subjects)
- Generates semantic theme summary

#### Stage 2: Creative Direction

The system:
- Maps product attributes to article context
- Generates 3 creative prompts that blend both:
  - Prompt 1: Direct integration (product as tool/object in scene)
  - Prompt 2: Ambient integration (product visible but not hero)
  - Prompt 3: Narrative integration (product as part of story)

#### Stage 3: Asset Generation

For each creative prompt:
- Generate a base banner design (1200x628 px recommended)
- Render vertical variant (300x600 px)
- Render icon variant (256x256 px)

#### Stage 4: Operator Review

The operator:
- Views 3 candidate banner designs
- Selects preferred design or rejects all and regenerates
- Chooses which format variants to export (square/vertical/icon)

#### Stage 5: Export

The system:
- Renders selected variants at production quality
- Packages as downloadable ZIP with metadata
- Returns ready-for-web PNG assets

### 6.3 Product Constraints

- **Maximum article length:** 50,000 characters (reduces processing cost)
- **Image format output:** PNG with transparency support
- **Target format variants:** square (1200x628), vertical (300x600), icon (256x256)
- **Generation time SLA:** < 3 minutes from submission to first preview
- **Regeneration limit:** Up to 2 re-picks per session (cost/time control)
- **No fallback generation:** If AI generation fails, surface error clearly

### 6.4 Success Criteria for Phase 5

- Article analysis extracts coherent themes
- Generated banners feel thematically connected to article + product
- All three format variants render without degradation
- End-to-end flow completes in < 5 minutes
- Operator can reject and regenerate with different creative direction
- Exported assets are production-ready (transparent backgrounds, correct dimensions)

## 7. Out of Scope (Intentional Deferrals)

- **A/B testing variants:** Operator can pick one design; A/B testing comes later
- **Animated banners:** MP4 or GIF output deferred; Phase 5 is static PNG only
- **Custom fonts:** Uses platform-standard typography; custom fonts future work
- **Brand compliance checking:** No AI moderation of brand safety (manual review only)
- **Programmatic placement:** No auto-placement to ad networks; manual export only
- **Analytics integration:** No tracking pixel or conversion monitoring

## 8. MVP Positioning

### In Relation to Phase 0-4

Phase 5 is **complementary, not replacement**:

- **Phase 0-4:** CAFAI video ad insertion (premium, immersive, complex)
- **Phase 5:** Dynamic website ads (quick, scalable, high-volume)
- **Together:** A complete "AI ad platform" serving both video and web channels

### Marketing Message

> "CAFAI now handles your full ad creative pipeline: contextual video insertion for premium streaming, and contextual banner generation for web. One platform, two channels, endless creative possibilities."

## 9. Success Metrics (Post-MVP)

- Time to asset creation (target: < 5 minutes vs. 2-4 weeks manual)
- Operator satisfaction with banner quality (target: 4+/5)
- Accepted designs on first attempt (target: 70%+)
- Average banners generated per user session (target: 3+)

## 10. Relationship to CAFAI Core

Phase 5 reuses CAFAI's fundamental strengths:

| Capability | Phase 0-4 (Video) | Phase 5 (Web) |
|------------|------------------|--------------|
| Content analysis | Video indexing | Article text analysis |
| Context extraction | Scene + character | Theme + mood |
| AI generation | Narrative video clip | Banner design |
| Product mapping | On-screen interaction | Visual integration |
| Output rendering | MP4 stitching | PNG export |
| Async orchestration | Job queue | Same queue pattern |

The same job orchestration, provider abstraction, and error handling patterns apply.

---

**Next Step:** See [02_System_Architecture_Document.md](02_System_Architecture_Document.md) for technical architecture and data flow.
