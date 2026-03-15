export const publicCopy = {
  brand: {
    name: "CAFAI",
    tagline: "Context-Aware Fused Ad Insertion",
    badge: "official demo",
  },
  nav: {
    home: "Home",
    gallery: "Gallery",
    upload: "Upload",
    about: "About",
  },
  landing: {
    eyebrow: "Official CAFAI demo",
    title: "CAFAI turns product insertion into a scene-aware, watchable cut.",
    lede:
      "CAFAI stands for Context-Aware Fused Ad Insertion. It analyzes a source video, finds a believable insertion moment, generates a short bridge clip, and stitches the result back into the scene as one previewable output.",
    primaryCta: "See the proof wall",
    secondaryCta: "Open the gallery",
    heroNoteTitle: "What CAFAI is",
    heroNote:
      "Context-Aware means the system reads the scene before inserting anything. Fused means the generated clip is stitched back into the original footage. Ad Insertion means the final output is a product moment that still feels native to the video.",
    heroStatsTitle: "Quick brag strip",
    galleryEyebrow: "Example garden",
    galleryTitle: "Pick a pink cut and look at the receipts.",
    galleryLede:
      "The gallery keeps every processed example in one place, while the home page stays focused on one hero example and the core explanation.",
    teaserEyebrow: "Processed videos",
    teaserTitle: "Only one featured example lives on the home page.",
    teaserLede:
      "The landing page stays focused on the strongest cut. The full gallery keeps every processed run together so the demo does not feel crowded.",
    teaserCta: "Open the full gallery",
    proofEyebrow: "Proof over buzzwords",
    proofTitle: "One featured scene. Four receipts right under it.",
    proofLede:
      "The proof rail makes it obvious what the source looked like, where the insert window lives, what was generated, and how the final cut holds together.",
    stepsEyebrow: "How it works",
    stepsTitle: "Cute on the outside, pipeline muscle underneath.",
    stepsLede:
      "The landing page stays flirty, but the explanation still names the real mechanics behind the demo.",
    ctaEyebrow: "Try your own footage",
    ctaTitle: "Bring your clip into the pink little block world.",
    ctaLede:
      "Upload one MP4, name the brand, and let the pipeline start reading scene rhythm, slotting an insert window, and preparing a reviewable preview.",
    ctaPrimary: "Start an upload",
    ctaSecondary: "Open the full gallery",
    aboutEyebrow: "Who built this",
    aboutTitle: "Two developers built the CAFAI demo and the workflow behind it.",
    aboutLede:
      "The About page is kept simple: who the two developers are and what each one focused on for the project.",
    aboutCta: "Meet the duo",
  },
  upload: {
    eyebrow: "Drop a new clip",
    title: "Hand over a video and let the CAFAI pipeline get busy.",
    lede:
      "Give the run a name, pick the brand, and drop in an MP4. The public flow stays clean while the operator tooling keeps working behind the curtain.",
    chips: ["One sweet upload step", "Proof-ready output", "Operator view stays available"],
    primaryCta: "Start the pretty pipeline",
    secondaryCta: "Back to home",
    productCta: "Upload a product",
    dropzoneTitle: "Source video",
    dropzoneHint: "Drag an MP4 here or tap to browse.",
    dropzoneSubhint: "Best for demo runs: H.264 MP4 with a clean scene change.",
    selectedFileLabel: "Selected file",
    resetLabel: "Pick another clip",
    statusTitle: "Pipeline status",
    statusQueued: "The upload is in, and the cute little render machine is spinning up.",
    statusCompleted: "Your preview is ready for a polished playback pass.",
    reviewLink: "Open studio review",
    previewLink: "Open preview theater",
  },
  about: {
    eyebrow: "About CAFAI",
    title: "The two developers behind the CAFAI demo.",
    lede: "A compact profile page for the two people who built the frontend and the pipeline flow.",
    cards: [
      {
        name: "Vlad",
        role: "Did everything except frontend and design choices.",
        bio: "Built the heavy-lifting parts of the project and carried the full technical backbone behind the demo.",
        avatar: "/about/vlad-tired-cat.jpg",
        github: "https://github.com/ch1kim0n1",
        githubLabel: "@ch1kim0n1",
      },
      {
        name: "Monika Jaqeli",
        role: "Design, frontend, picked videos for demo, was here for the vibes and Vlad's moral support.",
        bio:
          "Shaped the visual direction, handled the frontend presentation, chose the demo footage, and kept the project energy exactly where it needed to be.",
        avatar: "/about/monika-meme.jpg",
        github: "https://github.com/SuperLepeshka",
        githubLabel: "@SuperLepeshka",
      },
    ],
  },
} as const;
