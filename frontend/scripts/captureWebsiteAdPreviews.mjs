import fs from "node:fs/promises";
import path from "node:path";
import { chromium } from "playwright";

const repoRoot = path.resolve(process.cwd(), "..");
const publicDir = path.join(process.cwd(), "public", "website-ads");

const examples = [
  {
    id: "example1",
    url: "https://www.britannica.com/biography/Julius-Caesar-Roman-ruler",
    output: "site-preview.png",
  },
  {
    id: "example2",
    url: "https://www.forbes.com/profile/ted-turner/",
    output: "site-preview.png",
  },
  {
    id: "example3",
    url: "https://www.asianstudies.org/publications/eaa/archives/teaching-cultural-historical-and-religious-landscapes-with-the-anime-demon-slayer/",
    output: "site-preview.png",
  },
];

const browser = await chromium.launch({ headless: true });

try {
  for (const example of examples) {
    const exampleDir = path.join(publicDir, example.id);
    const bannerImage = path.join(exampleDir, "horizontal.png");
    const verticalImage = path.join(exampleDir, "vertical.png");
    const screenshotPath = path.join(exampleDir, example.output);

    const [bannerDataUrl, verticalDataUrl] = await Promise.all([
      fileToDataUrl(bannerImage),
      fileToDataUrl(verticalImage),
    ]);

    const context = await browser.newContext({
      viewport: { width: 1440, height: 1600 },
      deviceScaleFactor: 1,
    });

    const page = await context.newPage();
    await page.goto(example.url, { waitUntil: "domcontentloaded", timeout: 90_000 });
    await page.waitForTimeout(3_000);

    await page.addStyleTag({
      content: `
        [data-cafai-injected-ad],
        #onetrust-banner-sdk,
        .onetrust-pc-dark-filter,
        .fc-consent-root,
        .tp-modal,
        .tp-backdrop,
        .privacy-notice,
        .cookie-banner,
        .newsletter-signup,
        .modal-backdrop {
          display: none !important;
        }

        html, body {
          scroll-behavior: auto !important;
        }

        body {
          overflow-x: hidden !important;
        }

        .cafai-banner-slot {
          width: min(1200px, calc(100vw - 64px));
          margin: 24px auto 32px;
          padding: 0;
          border-radius: 18px;
          overflow: hidden;
          box-shadow: 0 18px 40px rgba(22, 12, 28, 0.28);
          border: 1px solid rgba(255, 255, 255, 0.65);
          background: white;
          position: relative;
          z-index: 8;
        }

        .cafai-banner-slot img {
          display: block;
          width: 100%;
          height: auto;
        }

        .cafai-banner-slot::before,
        .cafai-vertical-slot::before {
          content: "Sponsored preview";
          position: absolute;
          top: 10px;
          left: 10px;
          z-index: 2;
          padding: 4px 8px;
          border-radius: 999px;
          background: rgba(20, 16, 24, 0.78);
          color: white;
          font: 700 11px/1.2 system-ui, sans-serif;
          letter-spacing: 0.04em;
          text-transform: uppercase;
        }

        .cafai-vertical-slot {
          width: 300px;
          position: fixed;
          top: 148px;
          right: 18px;
          z-index: 9;
          border-radius: 18px;
          overflow: hidden;
          box-shadow: 0 18px 40px rgba(22, 12, 28, 0.28);
          border: 1px solid rgba(255, 255, 255, 0.65);
          background: white;
        }

        .cafai-vertical-slot img {
          display: block;
          width: 100%;
          height: auto;
        }
      `,
    });

    await page.evaluate(({ bannerSrc, verticalSrc }) => {
      const removeSelectors = [
        "iframe",
        "[aria-label*='advert']",
        "[id*='advert']",
        "[class*='advert']",
        "[id*='ad-']",
        "[class*='ad-']",
        "[id*='newsletter']",
        "[class*='newsletter']",
        "[role='dialog']",
      ];

      for (const selector of removeSelectors) {
        document.querySelectorAll(selector).forEach((node) => {
          const element = node;
          if (element instanceof HTMLElement && !element.closest("[data-cafai-preserve]")) {
            element.style.display = "none";
          }
        });
      }

      const banner = document.createElement("div");
      banner.className = "cafai-banner-slot";
      banner.dataset.cafaiInjectedAd = "banner";

      const bannerImg = document.createElement("img");
      bannerImg.src = bannerSrc;
      bannerImg.alt = "Injected banner ad preview";
      banner.appendChild(bannerImg);

      const vertical = document.createElement("aside");
      vertical.className = "cafai-vertical-slot";
      vertical.dataset.cafaiInjectedAd = "vertical";

      const verticalImg = document.createElement("img");
      verticalImg.src = verticalSrc;
      verticalImg.alt = "Injected vertical ad preview";
      vertical.appendChild(verticalImg);

      const insertionTarget =
        document.querySelector("main") ||
        document.querySelector("article") ||
        document.querySelector("#main-content") ||
        document.body.firstElementChild ||
        document.body;

      insertionTarget?.parentElement?.insertBefore(banner, insertionTarget);
      document.body.appendChild(vertical);
      window.scrollTo(0, 0);
    }, { bannerSrc: bannerDataUrl, verticalSrc: verticalDataUrl });

    await page.screenshot({ path: screenshotPath, fullPage: false });
    await context.close();
    console.log(`captured ${path.relative(repoRoot, screenshotPath)}`);
  }
} finally {
  await browser.close();
}

async function fileToDataUrl(filePath) {
  const extension = path.extname(filePath).toLowerCase();
  const mimeType = extension === ".jpg" || extension === ".jpeg" ? "image/jpeg" : "image/png";
  const buffer = await fs.readFile(filePath);
  return `data:${mimeType};base64,${buffer.toString("base64")}`;
}
