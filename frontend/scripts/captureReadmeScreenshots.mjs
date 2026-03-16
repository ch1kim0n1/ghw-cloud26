import fs from "node:fs/promises";
import path from "node:path";
import { chromium } from "playwright";

const frontendRoot = process.cwd();
const repoRoot = path.resolve(frontendRoot, "..");
const outputDir = path.join(repoRoot, "readme-assets");
const baseUrl = process.env.README_CAPTURE_BASE_URL ?? "http://127.0.0.1:4173";

const shots = [
  {
    name: "home.png",
    url: `${baseUrl}/`,
    viewport: { width: 1280, height: 800 },
  },
  {
    name: "upload.png",
    url: `${baseUrl}/upload`,
    viewport: { width: 1280, height: 800 },
  },
  {
    name: "gallery.png",
    url: `${baseUrl}/gallery`,
    viewport: { width: 1280, height: 800 },
  },
  {
    name: "about.png",
    url: `${baseUrl}/about`,
    viewport: { width: 1280, height: 800 },
  },
  {
    name: "proof-room.png",
    url: `${baseUrl}/`,
    viewport: { width: 1280, height: 900 },
    prepare: async (page) => {
      await page.locator("#proof-room").scrollIntoViewIfNeeded();
      await page.waitForTimeout(300);
    },
  },
];

await fs.mkdir(outputDir, { recursive: true });

const browser = await chromium.launch({ headless: true });

try {
  for (const shot of shots) {
    const page = await browser.newPage({
      viewport: shot.viewport,
      deviceScaleFactor: 1,
    });

    await page.goto(shot.url, { waitUntil: "networkidle" });
    await page.addStyleTag({
      content: `
        * {
          caret-color: transparent !important;
        }
      `,
    });

    if (shot.prepare) {
      await shot.prepare(page);
    }

    await page.screenshot({
      path: path.join(outputDir, shot.name),
      fullPage: false,
    });

    await page.close();
    console.log(`captured readme-assets/${shot.name}`);
  }
} finally {
  await browser.close();
}
