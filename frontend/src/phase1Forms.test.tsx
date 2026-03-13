import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { CampaignForm } from "./components/CampaignForm";
import { ProductForm } from "./components/ProductForm";

describe("Phase 1 forms", () => {
  beforeEach(() => {
    vi.stubGlobal("fetch", vi.fn());
  });

  afterEach(() => {
    cleanup();
    vi.unstubAllGlobals();
  });

  it("creates a product and renders it in the catalog", async () => {
    vi.mocked(fetch)
      .mockResolvedValueOnce({
        ok: true,
        json: async () => ({ products: [] }),
      } as Response)
      .mockResolvedValueOnce({
        ok: true,
        json: async () => ({
          id: "prod_1",
          name: "sparkling water",
          description: "light citrus",
          category: "beverage",
          context_keywords: ["drink", "water"],
          source_url: "https://example.com/water",
          created_at: "2026-03-13T00:00:00Z",
        }),
      } as Response);

    render(<ProductForm />);

    await screen.findByText("No products yet.");

    fireEvent.change(screen.getByLabelText("Name"), { target: { value: "sparkling water" } });
    fireEvent.change(screen.getByLabelText("Description"), { target: { value: "light citrus" } });
    fireEvent.change(screen.getByLabelText("Source URL"), { target: { value: "https://example.com/water" } });

    fireEvent.submit(screen.getByRole("button", { name: "Create product" }).closest("form") as HTMLFormElement);

    await screen.findByText("Created sparkling water.");
    expect(screen.getByText("sparkling water")).toBeInTheDocument();

    const postCall = vi.mocked(fetch).mock.calls[1];
    expect(postCall?.[0]).toContain("/api/products");
  });

  it("creates a campaign with an existing product", async () => {
    vi.mocked(fetch)
      .mockResolvedValueOnce({
        ok: true,
        json: async () => ({
          products: [
            {
              id: "prod_existing",
              name: "existing soda",
              created_at: "2026-03-13T00:00:00Z",
            },
          ],
        }),
      } as Response)
      .mockResolvedValueOnce({
        ok: true,
        json: async () => ({
          campaign_id: "camp_1",
          job_id: "job_1",
          product_id: "prod_existing",
          name: "Kitchen test",
          status: "queued",
          current_stage: "ready_for_analysis",
          video_filename: "clip.mp4",
          video_path: "tmp/uploads/campaigns/camp_1.mp4",
          source_fps: 23.976,
          duration_seconds: 601,
          target_ad_duration_seconds: 6,
          created_at: "2026-03-13T00:00:00Z",
        }),
      } as Response);

    render(<CampaignForm />);

    await screen.findByRole("option", { name: "existing soda" });

    fireEvent.change(screen.getByLabelText("Campaign name"), { target: { value: "Kitchen test" } });
    fireEvent.change(screen.getByLabelText("Source video (H.264 MP4)"), {
      target: { files: [new File(["video"], "clip.mp4", { type: "video/mp4" })] },
    });

    fireEvent.submit(screen.getByRole("button", { name: "Create campaign" }).closest("form") as HTMLFormElement);

    await screen.findByText("Campaign created");
    expect(screen.getByText("job_1")).toBeInTheDocument();
    expect(screen.getByText("queued")).toBeInTheDocument();

    const postCall = vi.mocked(fetch).mock.calls[1];
    const body = postCall?.[1]?.body;
    expect(body).toBeInstanceOf(FormData);
    expect((body as FormData).get("product_id")).toBe("prod_existing");
  });

  it("creates a campaign with an inline product when no products exist", async () => {
    vi.mocked(fetch)
      .mockResolvedValueOnce({
        ok: true,
        json: async () => ({ products: [] }),
      } as Response)
      .mockResolvedValueOnce({
        ok: true,
        json: async () => ({
          campaign_id: "camp_inline",
          job_id: "job_inline",
          product_id: "prod_inline",
          name: "Inline test",
          status: "queued",
          current_stage: "ready_for_analysis",
          video_filename: "inline.mp4",
          video_path: "tmp/uploads/campaigns/camp_inline.mp4",
          source_fps: 24,
          duration_seconds: 601,
          target_ad_duration_seconds: 6,
          created_at: "2026-03-13T00:00:00Z",
        }),
      } as Response);

    render(<CampaignForm />);

    await waitFor(() => {
      expect(screen.getByRole("radio", { name: "Create inline product" })).toBeChecked();
    });

    fireEvent.change(screen.getByLabelText("Campaign name"), { target: { value: "Inline test" } });
    fireEvent.change(screen.getByLabelText("Source video (H.264 MP4)"), {
      target: { files: [new File(["video"], "inline.mp4", { type: "video/mp4" })] },
    });
    fireEvent.change(screen.getByLabelText("Product name"), { target: { value: "inline soda" } });
    fireEvent.change(screen.getByLabelText("Source URL"), {
      target: { value: "https://example.com/inline-soda" },
    });

    fireEvent.submit(screen.getByRole("button", { name: "Create campaign" }).closest("form") as HTMLFormElement);

    await screen.findByText("Campaign created");

    const postCall = vi.mocked(fetch).mock.calls[1];
    const body = postCall?.[1]?.body as FormData;
    expect(body.get("product_name")).toBe("inline soda");
    expect(body.get("product_source_url")).toBe("https://example.com/inline-soda");
  });
});
