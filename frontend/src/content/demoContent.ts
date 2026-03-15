export type DemoExample = {
  id: string;
  featured: boolean;
  label: string;
  displayName: string;
  shortTag: string;
  title: string;
  heroBlurb: string;
  summary: string;
  scene: string;
  jobId: string;
  sourceDurationSeconds: number;
  insertStartSeconds: number;
  insertedDurationSeconds: number;
  previewDurationSeconds: number;
  anchorFrames: string;
  selectedWindow: string;
  finalVideo: string;
  finalPoster: string;
  generatedPreview: string;
  startFrame: string;
  stopFrame: string;
  proofLabels: {
    original: string;
    window: string;
    bridge: string;
    final: string;
  };
  palette: {
    sky: string;
    panel: string;
    accent: string;
    border: string;
    shadow: string;
    grass: string;
  };
  decorAssetIds: string[];
};

export const demoExamples: DemoExample[] = [
  {
    id: "example1",
    featured: false,
    label: "Bike Bloom",
    displayName: "Bike Bloom Reveal",
    shortTag: "late-scene glow-up",
    title: "Outdoor reveal with a late-scene handoff",
    heroBlurb: "A breezy outdoor handoff that keeps the energy high before the branded moment arrives.",
    summary: "A bicycle scene where the inserted branded moment lands late, preserving flow and camera energy.",
    scene: "Outdoor bicycle sequence",
    jobId: "job_6678aff9-e05f-49c9-b4ee-1ffd0a9a0863",
    sourceDurationSeconds: 59.526,
    insertStartSeconds: 41.708,
    insertedDurationSeconds: 4.972,
    previewDurationSeconds: 64.498,
    anchorFrames: "1250 -> 1300",
    selectedWindow: "41.708s -> 43.377s",
    finalVideo: "/demo/example1-final.mp4",
    finalPoster: "/demo/example1-final-poster.png",
    generatedPreview: "/demo/example1-generated.gif",
    startFrame: "/demo/example1-start-frame.png",
    stopFrame: "/demo/example1-stop-frame.png",
    proofLabels: {
      original: "Original scene beat",
      window: "Late insert window",
      bridge: "Generated brand bridge",
      final: "Final stitched reveal",
    },
    palette: {
      sky: "#ffe7f4",
      panel: "#fff5fb",
      accent: "#ff7cae",
      border: "#8e4765",
      shadow: "#d05382",
      grass: "#d6f4bf",
    },
    decorAssetIds: ["cloud", "heart", "star"],
  },
  {
    id: "example2",
    featured: false,
    label: "Desk Darling",
    displayName: "Desk Darling Bridge",
    shortTag: "talking-head gloss",
    title: "Talking-head scene with a seamless branded bridge",
    heroBlurb: "A tidy desk scene that proves the inserted moment can sit inside a calm camera setup without looking pasted on.",
    summary:
      "A desk-side talking-head example that shows the inserted moment without falling back to a conventional cutaway ad.",
    scene: "Desk-side talking head",
    jobId: "job_9de1cbb7-ec84-4e2c-99f7-0d2dc6f21e0a",
    sourceDurationSeconds: 59.002,
    insertStartSeconds: 20.5,
    insertedDurationSeconds: 6.535,
    previewDurationSeconds: 65.537,
    anchorFrames: "615 -> 630",
    selectedWindow: "20.5s -> 21.0s",
    finalVideo: "/demo/example2-final.mp4",
    finalPoster: "/demo/example2-final-poster.png",
    generatedPreview: "/demo/example2-generated.gif",
    startFrame: "/demo/start-frame.png",
    stopFrame: "/demo/stop-frame.png",
    proofLabels: {
      original: "Original desk beat",
      window: "Mid-scene insert window",
      bridge: "Generated desk bridge",
      final: "Final stitched talk-through",
    },
    palette: {
      sky: "#fff0f7",
      panel: "#fff9fd",
      accent: "#ff8dc2",
      border: "#90506d",
      shadow: "#cf608d",
      grass: "#ddf8c9",
    },
    decorAssetIds: ["bow", "flower", "cloud"],
  },
  {
    id: "example3",
    featured: true,
    label: "Pixel Pop",
    displayName: "Pixel Pop Energy",
    shortTag: "anime desk sparkle",
    title: "Streamer close-up with an early energy-drink insert",
    heroBlurb:
      "The cutest example in the set: character-focused framing, a bright product beat, and proof that the brand moment can arrive early without wrecking the vibe.",
    summary:
      "An anime desk scene where the branded drink moment lands right after the opening beat without dropping the character-focused framing.",
    scene: "Streamer desk close-up",
    jobId: "job_592fecd7-ff36-4beb-acba-170ce0f16107",
    sourceDurationSeconds: 82.199,
    insertStartSeconds: 7.9,
    insertedDurationSeconds: 7.042,
    previewDurationSeconds: 88.507,
    anchorFrames: "237 -> 259",
    selectedWindow: "7.9s -> 8.633s",
    finalVideo: "/demo/example3-final.mp4",
    finalPoster: "/demo/example3-final-poster.png",
    generatedPreview: "/demo/example3-generated.gif",
    startFrame: "/demo/example3-start-frame.png",
    stopFrame: "/demo/example3-stop-frame.png",
    proofLabels: {
      original: "Original opening beat",
      window: "Early insert window",
      bridge: "Generated drink bridge",
      final: "Final stitched idol cut",
    },
    palette: {
      sky: "#ffe4f0",
      panel: "#fff3f9",
      accent: "#ff6da5",
      border: "#87445f",
      shadow: "#c14b79",
      grass: "#d4f0b0",
    },
    decorAssetIds: ["cloud", "bow", "flower", "star"],
  },
];

export const featuredDemoExample =
  demoExamples.find((example) => example.featured) ?? demoExamples[demoExamples.length - 1];

export const latestDemoExample = featuredDemoExample;

export const heroStats = [
  { value: String(demoExamples.length), label: "polished examples" },
  { value: "4", label: "proof checkpoints" },
  { value: "under 10s", label: "to understand the pitch" },
];

export const demoSteps = [
  "Read the scene rhythm instead of forcing a clunky ad break.",
  "Pick a believable insert window and anchor it to real frames.",
  "Generate a short branded bridge that fits the tone of the scene.",
  "Render a preview that operators can review or download right away.",
];

export const proofPoints = [
  {
    title: "Finished cut first",
    body: "The landing page opens with the stitched result so the audience sees the value before they see any process.",
  },
  {
    title: "Receipts on display",
    body: "Original frame, insert window, generated bridge, and final cut are split into a clean proof rail instead of a messy screenshot pile.",
  },
  {
    title: "Cute, not vague",
    body: "The language stays playful, but it still names the real mechanics: scene rhythm, anchor frames, review flow, and preview rendering.",
  },
];

export const studioLinks = [
  { label: "Operator catalog", to: "/products" },
  { label: "Create campaign", to: "/campaigns/new" },
  { label: "Open live job", to: `/jobs/${latestDemoExample.jobId}` },
];

export function buildTimelineSegments(example: DemoExample) {
  return [
    {
      label: "Source before insertion",
      seconds: example.insertStartSeconds,
      tone: "base" as const,
    },
    {
      label: "Generated bridge",
      seconds: example.insertedDurationSeconds,
      tone: "inserted" as const,
    },
    {
      label: "Source after insertion",
      seconds: example.sourceDurationSeconds - example.insertStartSeconds,
      tone: "base" as const,
    },
  ];
}

export function buildTimelineScale(sourceDurationSeconds: number) {
  const roundedMax = Math.round(sourceDurationSeconds);
  const steps = [0, 10, 20, 30, 40, 50, roundedMax].filter((step, index, values) => values.indexOf(step) === index);
  return steps.map((step) => `${step}s`);
}
