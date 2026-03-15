export type DemoExample = {
  id: string;
  label: string;
  title: string;
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
};

export const demoExamples: DemoExample[] = [
  {
    id: "example1",
    label: "Example 01",
    title: "Outdoor reveal with a late-scene handoff",
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
  },
  {
    id: "example2",
    label: "Example 02",
    title: "Talking-head scene with a seamless branded bridge",
    summary: "A desk-side talking-head example that shows the inserted moment without falling back to a conventional cutaway ad.",
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
  },
  {
    id: "example3",
    label: "Example 03",
    title: "Streamer close-up with an early energy-drink insert",
    summary: "An anime desk scene where the branded drink moment lands right after the opening beat without dropping the character-focused framing.",
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
  },
];

export const latestDemoExample = demoExamples.find((example) => example.id === "example2") ?? demoExamples[demoExamples.length - 1];

export const heroStats = [
  { value: String(demoExamples.length), label: "completed demo cuts" },
  { value: "1", label: "scene-aware insert per example" },
  { value: "under 10s", label: "to understand the value" },
];

export const demoSteps = [
  "Read the scene rhythm instead of forcing an ad break.",
  "Choose a believable insertion window with anchor frames.",
  "Generate a short branded bridge clip that fits the scene.",
  "Export a preview that can be reviewed live with one click.",
];

export const proofPoints = [
  {
    title: "Product story first",
    body: "The upgraded demo leads with the stitched output, not the control plane, so judges immediately see the value proposition.",
  },
  {
    title: "Temporal proof, not abstract claims",
    body: "Timelines, anchor frames, and generated clips explain exactly where the branded moment lives inside the original scene.",
  },
  {
    title: "Real pipeline underneath",
    body: "The studio tools remain available for operators, but they no longer dominate the presentation surface.",
  },
];

export const studioLinks = [
  { label: "Studio dashboard", to: "/products" },
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
  const steps = [0, 10, 20, 30, 40, 50, Math.round(sourceDurationSeconds)];
  return steps.map((step) => `${step}s`);
}
