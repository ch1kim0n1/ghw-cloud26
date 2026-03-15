import { cleanup, fireEvent, render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { UploadPage } from "./pages/UploadPage";

describe("UploadPage", () => {
  beforeEach(() => {
    vi.stubGlobal("fetch", vi.fn());
  });

  afterEach(() => {
    cleanup();
    vi.unstubAllGlobals();
  });

  it("uploads a video, auto-starts analysis, and shows the studio-ready progress state", async () => {
    vi.mocked(fetch).mockImplementation(async (input: RequestInfo | URL, init?: RequestInit) => {
      const url = String(input);

      if (url.includes("/api/campaigns") && init?.method === "POST") {
        return {
          ok: true,
          json: async () => ({
            campaign_id: "camp_1",
            job_id: "job_1",
            product_id: "prod_inline",
            name: "Cherry pixel dream",
            status: "queued",
            current_stage: "ready_for_analysis",
            video_filename: "clip.mp4",
            video_path: "tmp/uploads/campaigns/camp_1.mp4",
            target_ad_duration_seconds: 6,
            created_at: "2026-03-13T00:00:00Z",
          }),
        } as Response;
      }

      if (url.includes("/api/jobs/job_1/start-analysis") && init?.method === "POST") {
        return {
          ok: true,
          json: async () => ({
            job_id: "job_1",
            status: "analyzing",
            current_stage: "analysis_submission",
            message: "analysis started",
          }),
        } as Response;
      }

      if (url.includes("/api/jobs/job_1/preview")) {
        return {
          ok: false,
          status: 404,
          json: async () => ({
            error: "preview not found",
            error_code: "RESOURCE_NOT_FOUND",
          }),
        } as Response;
      }

      if (url.includes("/api/jobs/job_1")) {
        return {
          ok: true,
          json: async () => ({
            id: "job_1",
            campaign_id: "camp_1",
            status: "analyzing",
            current_stage: "slot_selection",
            progress_percent: 42,
            selected_slot_id: null,
            error_message: null,
            error_code: null,
            created_at: "2026-03-13T00:00:00Z",
            started_at: "2026-03-13T00:01:00Z",
            completed_at: null,
          }),
        } as Response;
      }

      return {
        ok: false,
        status: 404,
        json: async () => ({ error: "not found", error_code: "RESOURCE_NOT_FOUND" }),
      } as Response;
    });

    render(
      <MemoryRouter>
        <UploadPage />
      </MemoryRouter>,
    );

    fireEvent.change(screen.getByLabelText("Campaign name"), { target: { value: "Cherry pixel dream" } });
    fireEvent.change(screen.getByLabelText("Brand / Product name"), { target: { value: "Cherry Pop" } });
    fireEvent.change(screen.getByLabelText("Source video"), {
      target: { files: [new File(["video"], "clip.mp4", { type: "video/mp4" })] },
    });

    expect(screen.getByText("clip.mp4")).toBeInTheDocument();

    fireEvent.submit(screen.getByRole("button", { name: "Start the pretty pipeline" }).closest("form") as HTMLFormElement);

    await screen.findByText("Pipeline status");
    expect(screen.getByText("Open studio review")).toBeInTheDocument();
    expect(screen.getByText("slot_selection")).toBeInTheDocument();
    expect(screen.getByText("42%")).toBeInTheDocument();

    const createCall = vi.mocked(fetch).mock.calls.find(([url]) => String(url).includes("/api/campaigns"));
    const body = createCall?.[1]?.body as FormData;
    expect(body.get("product_name")).toBe("Cherry Pop");
    expect(body.get("target_ad_duration_seconds")).toBe("6");

    expect(
      vi.mocked(fetch).mock.calls.some(([url]) => String(url).includes("/api/jobs/job_1/start-analysis")),
    ).toBe(true);
  });
});
