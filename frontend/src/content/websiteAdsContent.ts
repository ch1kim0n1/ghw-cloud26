import { assetUrl } from "../utils/assetUrl";

export const websiteAdsContent = {
  chips: ["Stable Diffusion backend", "Banner + vertical output", "Saved locally in the main demo"],
  examples: [
    {
      id: "example1",
      label: "Example 1",
      title: "Julius Caesar reference article",
      url: "https://www.britannica.com/biography/Julius-Caesar-Roman-ruler",
      previewImage: assetUrl("website-ads/example1/injected-preview.png"),
      note: "Captured article page with both placements moved into the empty media block so the article copy stays untouched.",
      preview: {
        publication: "History Weekly",
        section: "Ancient Rome",
        headline: "Julius Caesar still dominates the Roman imagination.",
        dek:
          "A long-form profile revisits Caesar's rise, public image, military campaigns, and why the mythology still pulls modern readers in.",
        byline: "Previewing on an editorial article shell",
        body: [
          "Roman political history is crowded with powerful names, but Julius Caesar still behaves like a gravitational center. His biography holds together conquest, theater, propaganda, and the collapse of republican norms in one story that never quite stops feeling contemporary.",
          "That is exactly the kind of article context a static ad can borrow from. The surrounding page already has marble, bronze, empire, spectacle, and ambition in the reader's head before the product creative ever appears.",
          "The preview below simulates a top-of-article banner and a right-rail vertical unit so you can judge whether the generated art feels placed rather than floating in a vacuum.",
        ],
      },
    },
    {
      id: "example2",
      label: "Example 2",
      title: "Ted Turner reference article",
      url: "https://www.forbes.com/profile/ted-turner/",
      previewImage: assetUrl("website-ads/example2/injected-preview.png"),
      note: "Captured Forbes profile page with a masthead banner in the top whitespace and the vertical unit moved into the right-side empty rail.",
      preview: {
        publication: "Business Ledger",
        section: "Forbes Profile",
        headline: "Ted Turner remains one of the most recognizable media moguls in the modern business canon.",
        dek:
          "A profile page like this frames Turner through ownership, media scale, legacy, and long-tail influence across broadcasting and philanthropy.",
        byline: "Previewing on a business-profile article shell",
        body: [
          "A business profile page gives the reader a different context than a historical biography. The signals here are wealth, media, legacy, influence, and reputation rather than spectacle or ancient drama.",
          "That changes what a good ad placement feels like. The page can support a sharper, cleaner visual language and still let the creative feel native to the surrounding editorial surface.",
          "This preview uses your manually added Example 2 banner and vertical creative to simulate how those assets would read at the top of a profile page and in a sidebar slot.",
        ],
      },
    },
    {
      id: "example3",
      label: "Example 3",
      title: "Demon Slayer cultural landscapes reference article",
      url: "https://www.asianstudies.org/publications/eaa/archives/teaching-cultural-historical-and-religious-landscapes-with-the-anime-demon-slayer/",
      previewImage: assetUrl("website-ads/example3/injected-preview.png"),
      note: "Fresh capture of the article page with the banner placed in the site chrome and the vertical unit placed in unused sidebar space.",
      preview: {
        publication: "Culture Review",
        section: "Anime and Landscape",
        headline: "Demon Slayer opens a path into cultural, historical, and religious landscapes.",
        dek:
          "An academic teaching article like this frames anime through place, symbolism, memory, and interpretation rather than pure fandom alone.",
        byline: "Previewing on a culture-and-education article shell",
        body: [
          "This page carries a very different editorial signal from the first two examples. It is slower, more interpretive, and rooted in cultural reading rather than celebrity or executive profile energy.",
          "That gives the ad placement preview a different test. The creative needs to feel visually compatible with a thoughtful article page while still remaining bright enough to act like an ad unit.",
          "This mock uses your Example 3 banner and vertical assets to simulate a top placement and a sidebar placement on a culture-and-education style article layout.",
        ],
      },
    },
  ],
  styleOptions: [
    "playful editorial",
    "clean minimal",
    "retro-futurist",
    "luxury magazine",
    "soft pastel",
  ],
} as const;
