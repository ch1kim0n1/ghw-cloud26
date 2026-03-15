import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { afterEach, describe, expect, it, vi } from "vitest";
import { JobPage } from "./pages/JobPage";

describe("Phase 3 job page", () => {
  afterEach(() => {
    cleanup();
    vi.unstubAllGlobals();
  });

  it("supports slot review, line review, and generation from the dashboard", async () => {
    const state: {
      job: Record<string, unknown>;
      slots: Array<Record<string, unknown>>;
      logs: Array<Record<string, unknown>>;
    } = {
      job: {
        id: "job_1",
        campaign_id: "camp_1",
        status: "queued",
        current_stage: "ready_for_analysis",
        progress_percent: 0,
        selected_slot_id: null,
        error_message: null,
        error_code: null,
        created_at: "2026-03-13T00:00:00Z",
        started_at: null,
        completed_at: null,
        metadata: {
          source_fps: 24,
          duration_seconds: 900,
          repick_count: 0,
          rejected_slot_ids: [],
          top_slot_ids: [],
        },
      },
      slots: [],
      logs: [],
    };

    vi.stubGlobal(
      "fetch",
      vi.fn(async (input: RequestInfo | URL, init?: RequestInit) => {
        const url = String(input);

        if (url.includes("/api/health")) {
          return {
            ok: true,
            json: async () => ({
              status: "healthy",
              timestamp: "2026-03-15T00:00:00Z",
              version: "0.1.0",
              provider_profile: "azure",
              audit: {
                enabled: true,
                status: "healthy",
                details: "notion audit sink connected",
              },
            }),
          } as Response;
        }

        if (url.includes("/api/jobs/job_1/start-analysis") && init?.method === "POST") {
          state.job = {
            ...state.job,
            status: "analyzing",
            current_stage: "slot_selection",
            progress_percent: 40,
            started_at: "2026-03-13T00:01:00Z",
          };
          state.slots = [
            {
              id: "slot_1",
              rank: 1,
              scene_id: "scene_1",
              anchor_start_frame: 120,
              anchor_end_frame: 121,
              source_fps: 24,
              quiet_window_seconds: 4.2,
              score: 0.91,
              reasoning: "top ranked candidate",
              status: "proposed",
            },
            {
              id: "slot_2",
              rank: 2,
              scene_id: "scene_2",
              anchor_start_frame: 240,
              anchor_end_frame: 241,
              source_fps: 24,
              quiet_window_seconds: 3.8,
              score: 0.83,
              reasoning: "second ranked candidate",
              status: "proposed",
            },
          ];
          state.logs = [
            {
              timestamp: "2026-03-13T00:01:00Z",
              event_type: "stage_started",
              stage_name: "analysis_submission",
              message: "analysis started",
            },
          ];

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

        if (url.includes("/api/jobs/job_1/slots/slot_1/select") && init?.method === "POST") {
          state.job = {
            ...state.job,
            current_stage: "line_review",
            selected_slot_id: "slot_1",
          };
          state.slots = state.slots.map((slot) =>
            slot.id === "slot_1"
              ? {
                  ...slot,
                  status: "selected",
                  suggested_product_line: "I grabbed this sparkling water earlier.",
                }
              : slot,
          );
          state.logs = [
            ...state.logs,
            {
              timestamp: "2026-03-13T00:02:00Z",
              event_type: "slot_selected",
              stage_name: "line_review",
              message: "slot selected and product line prepared",
            },
          ];

          return {
            ok: true,
            json: async () => ({
              job_id: "job_1",
              slot_id: "slot_1",
              status: "analyzing",
              current_stage: "line_review",
              slot_status: "selected",
              suggested_product_line: "I grabbed this sparkling water earlier.",
              message: "slot selected and product line prepared",
            }),
          } as Response;
        }

        if (url.includes("/api/jobs/job_1/slots/slot_1/generate") && init?.method === "POST") {
          state.job = {
            ...state.job,
            status: "generating",
            current_stage: "generation_poll",
            progress_percent: 80,
          };
          state.slots = state.slots.map((slot) =>
            slot.id === "slot_1"
              ? {
                  ...slot,
                  status: "generated",
                  product_line_mode: "operator",
                  final_product_line: "I picked up this sparkling water earlier.",
                  generated_clip_path: "tmp/artifacts/job_1/slot_1.mp4",
                  generated_audio_path: "tmp/artifacts/job_1/slot_1.wav",
                }
              : slot,
          );
          state.logs = [
            ...state.logs,
            {
              timestamp: "2026-03-13T00:03:00Z",
              event_type: "stage_completed",
              stage_name: "generation_poll",
              message: "cafai generation complete",
            },
          ];

          return {
            ok: true,
            json: async () => ({
              job_id: "job_1",
              slot_id: "slot_1",
              status: "generating",
              current_stage: "generation_submission",
              slot_status: "generating",
              message: "cafai generation started",
            }),
          } as Response;
        }

        if (url.includes("/api/jobs/job_1/logs")) {
          return {
            ok: true,
            json: async () => ({ job_id: "job_1", logs: state.logs }),
          } as Response;
        }

        if (url.includes("/api/jobs/job_1/slots")) {
          return {
            ok: true,
            json: async () => ({ job_id: "job_1", slots: state.slots }),
          } as Response;
        }

        if (url.includes("/api/jobs/job_1")) {
          return {
            ok: true,
            json: async () => state.job,
          } as Response;
        }

        return {
          ok: false,
          status: 404,
          json: async () => ({ error: "not found", error_code: "RESOURCE_NOT_FOUND" }),
        } as Response;
      }),
    );

    render(
      <MemoryRouter initialEntries={["/jobs/job_1"]}>
        <Routes>
          <Route path="/jobs/:jobId" element={<JobPage />} />
        </Routes>
      </MemoryRouter>,
    );

    await waitFor(() => {
      expect(screen.getAllByText("queued").length).toBeGreaterThan(0);
    });

    fireEvent.click(screen.getByRole("button", { name: "Start analysis" }));

    await screen.findByText("top ranked candidate");
    fireEvent.click(screen.getAllByRole("button", { name: "Select" })[0]);

    await screen.findByText("Product Line Review");
    expect(screen.getByDisplayValue("I grabbed this sparkling water earlier.")).toBeInTheDocument();

    fireEvent.click(screen.getByRole("radio", { name: "Operator edit" }));
    fireEvent.change(screen.getByRole("textbox", { name: "Operator line" }), {
      target: { value: "I picked up this sparkling water earlier." },
    });
    fireEvent.click(screen.getByRole("button", { name: "Start generation" }));

    await screen.findByText("Generation complete.");
    expect(screen.getByText("tmp/artifacts/job_1/slot_1.mp4")).toBeInTheDocument();
    expect(screen.getByText("stage_completed")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Render preview" })).toBeInTheDocument();
  });

  it("shows preview open and download actions once rendering completes", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(async (input: RequestInfo | URL) => {
        const url = String(input);

        if (url.includes("/api/health")) {
          return {
            ok: true,
            json: async () => ({
              status: "healthy",
              timestamp: "2026-03-15T00:00:00Z",
              version: "0.1.0",
              provider_profile: "azure",
              audit: {
                enabled: true,
                status: "healthy",
                details: "notion audit sink connected",
              },
            }),
          } as Response;
        }

        if (url.includes("/api/jobs/job_2/logs")) {
          return {
            ok: true,
            json: async () => ({ job_id: "job_2", logs: [] }),
          } as Response;
        }

        if (url.includes("/api/jobs/job_2/slots")) {
          return {
            ok: true,
            json: async () => ({
              job_id: "job_2",
              slots: [
                {
                  id: "slot_1",
                  rank: 1,
                  scene_id: "scene_1",
                  anchor_start_frame: 120,
                  anchor_end_frame: 121,
                  source_fps: 24,
                  quiet_window_seconds: 4.2,
                  score: 0.91,
                  reasoning: "top ranked candidate",
                  status: "generated",
                  generated_clip_path: "tmp/artifacts/job_2/slot_1.mp4",
                },
              ],
            }),
          } as Response;
        }

        if (url.includes("/api/jobs/job_2/preview")) {
          return {
            ok: true,
            json: async () => ({
              id: "preview_1",
              job_id: "job_2",
              slot_id: "slot_1",
              status: "completed",
              output_video_path: "tmp/previews/job_2_preview.mp4",
              download_path: "/api/jobs/job_2/preview/download",
              duration_seconds: 906,
              render_retry_count: 0,
              created_at: "2026-03-13T00:00:00Z",
              completed_at: "2026-03-13T00:05:00Z",
              artifact_manifest: {},
              render_metrics: {},
            }),
          } as Response;
        }

        if (url.includes("/api/jobs/job_2")) {
          return {
            ok: true,
            json: async () => ({
              id: "job_2",
              campaign_id: "camp_2",
              status: "completed",
              current_stage: "render_poll",
              progress_percent: 100,
              selected_slot_id: "slot_1",
              error_message: null,
              error_code: null,
              created_at: "2026-03-13T00:00:00Z",
              started_at: "2026-03-13T00:01:00Z",
              completed_at: "2026-03-13T00:05:00Z",
              metadata: {
                source_fps: 24,
                duration_seconds: 900,
                repick_count: 0,
                rejected_slot_ids: [],
                top_slot_ids: [],
              },
            }),
          } as Response;
        }

        return {
          ok: false,
          status: 404,
          json: async () => ({ error: "not found", error_code: "RESOURCE_NOT_FOUND" }),
        } as Response;
      }),
    );

    render(
      <MemoryRouter initialEntries={["/jobs/job_2"]}>
        <Routes>
          <Route path="/jobs/:jobId" element={<JobPage />} />
        </Routes>
      </MemoryRouter>,
    );

    expect(await screen.findByText("Preview status: completed")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Open preview" })).toHaveAttribute("href", "/jobs/job_2/preview");
    expect(screen.getByRole("link", { name: "Download preview" })).toHaveAttribute(
      "href",
      "http://localhost:8080/api/jobs/job_2/preview/download",
    );
  });

  it("shows manual selection controls when no auto slot is available", async () => {
    const state: {
      job: Record<string, unknown>;
      slots: Array<Record<string, unknown>>;
      logs: Array<Record<string, unknown>>;
    } = {
      job: {
        id: "job_3",
        campaign_id: "camp_3",
        status: "analyzing",
        current_stage: "slot_selection",
        progress_percent: 40,
        selected_slot_id: null,
        error_message: "no suitable auto slot found; manual selection available",
        error_code: "NO_SUITABLE_SLOT_FOUND",
        created_at: "2026-03-13T00:00:00Z",
        started_at: "2026-03-13T00:01:00Z",
        completed_at: null,
        metadata: {
          source_fps: 24,
          duration_seconds: 900,
          content_language: "ru",
          repick_count: 0,
          rejected_slot_ids: [],
          top_slot_ids: [],
        },
      },
      slots: [],
      logs: [],
    };

    vi.stubGlobal(
      "fetch",
      vi.fn(async (input: RequestInfo | URL, init?: RequestInit) => {
        const url = String(input);

        if (url.includes("/api/health")) {
          return {
            ok: true,
            json: async () => ({
              status: "healthy",
              timestamp: "2026-03-15T00:00:00Z",
              version: "0.1.0",
              provider_profile: "azure",
              audit: {
                enabled: true,
                status: "healthy",
                details: "notion audit sink connected",
              },
            }),
          } as Response;
        }

        if (url.includes("/api/jobs/job_3/slots/manual-select") && init?.method === "POST") {
          state.job = {
            ...state.job,
            current_stage: "line_review",
            error_message: null,
            error_code: null,
            selected_slot_id: "slot_manual",
          };
          state.slots = [
            {
              id: "slot_manual",
              rank: 0,
              scene_id: "scene_1",
              anchor_start_frame: 48,
              anchor_end_frame: 144,
              source_fps: 24,
              quiet_window_seconds: 4,
              score: 0.61,
              reasoning: "manual selection by operator",
              status: "selected",
              suggested_product_line: "Я возьму эту бутылку на секунду.",
              metadata: {
                manual: true,
                manual_start_seconds: 2,
                manual_end_seconds: 6,
              },
            },
          ];

          return {
            ok: true,
            json: async () => ({
              job_id: "job_3",
              slot_id: "slot_manual",
              status: "analyzing",
              current_stage: "line_review",
              slot_status: "selected",
              suggested_product_line: "Я возьму эту бутылку на секунду.",
              manual: true,
              message: "manual slot selected and product line prepared",
            }),
          } as Response;
        }

        if (url.includes("/api/jobs/job_3/logs")) {
          return {
            ok: true,
            json: async () => ({ job_id: "job_3", logs: state.logs }),
          } as Response;
        }

        if (url.includes("/api/jobs/job_3/slots")) {
          return {
            ok: true,
            json: async () => ({ job_id: "job_3", slots: state.slots }),
          } as Response;
        }

        if (url.includes("/api/jobs/job_3")) {
          return {
            ok: true,
            json: async () => state.job,
          } as Response;
        }

        return {
          ok: false,
          status: 404,
          json: async () => ({ error: "not found", error_code: "RESOURCE_NOT_FOUND" }),
        } as Response;
      }),
    );

    render(
      <MemoryRouter initialEntries={["/jobs/job_3"]}>
        <Routes>
          <Route path="/jobs/:jobId" element={<JobPage />} />
        </Routes>
      </MemoryRouter>,
    );

    await screen.findByText("Detected content language: RU");
    expect(
      screen.getByText("Automatic ranking found no suitable slot. Manual selection is the primary recovery path."),
    ).toBeInTheDocument();

    fireEvent.change(screen.getByLabelText("Start seconds"), { target: { value: "2" } });
    fireEvent.change(screen.getByLabelText("End seconds"), { target: { value: "6" } });
    fireEvent.click(screen.getByRole("button", { name: "Select manual slot" }));

    await screen.findByText("Product Line Review");
    expect(screen.getByDisplayValue("Я возьму эту бутылку на секунду.")).toBeInTheDocument();
  });

  it("imports a locally generated clip into the selected slot", async () => {
    const state: {
      job: Record<string, unknown>;
      slots: Array<Record<string, unknown>>;
      logs: Array<Record<string, unknown>>;
      preview: Record<string, unknown> | null;
    } = {
      job: {
        id: "job_4",
        campaign_id: "camp_4",
        status: "analyzing",
        current_stage: "line_review",
        progress_percent: 40,
        selected_slot_id: "slot_4",
        error_message: null,
        error_code: null,
        created_at: "2026-03-13T00:00:00Z",
        started_at: "2026-03-13T00:01:00Z",
        completed_at: null,
        metadata: {
          source_fps: 30,
          duration_seconds: 59,
          content_language: "en",
        },
      },
      slots: [
        {
          id: "slot_4",
          rank: 1,
          scene_id: "scene_4",
          anchor_start_frame: 615,
          anchor_end_frame: 630,
          source_fps: 30,
          quiet_window_seconds: 4,
          score: 0.77,
          reasoning: "selected demo slot",
          status: "selected",
          suggested_product_line: "Pass me a Pepsi for this segment.",
        },
      ],
      logs: [],
      preview: null,
    };

    vi.stubGlobal(
      "fetch",
      vi.fn(async (input: RequestInfo | URL, init?: RequestInit) => {
        const url = String(input);

        if (url.includes("/api/jobs/job_4/slots/manual-import") && init?.method === "POST") {
          state.job = {
            ...state.job,
            status: "generating",
            current_stage: "generation_poll",
            progress_percent: 80,
          };
          state.slots = state.slots.map((slot) =>
            slot.id === "slot_4"
              ? {
                  ...slot,
                  status: "generated",
                  generated_clip_path: "/tmp/example2/generated.mp4",
                }
              : slot,
          );

          return {
            ok: true,
            json: async () => ({
              job_id: "job_4",
              slot_id: "slot_4",
              status: "generating",
              current_stage: "generation_poll",
              slot_status: "generated",
              generated_clip_path: "/tmp/example2/generated.mp4",
              manual: true,
              message: "manual generated clip imported",
            }),
          } as Response;
        }

        if (url.includes("/api/jobs/job_4/preview")) {
          return {
            ok: true,
            json: async () => state.preview,
          } as Response;
        }

        if (url.includes("/api/jobs/job_4/logs")) {
          return {
            ok: true,
            json: async () => ({ job_id: "job_4", logs: state.logs }),
          } as Response;
        }

        if (url.includes("/api/jobs/job_4/slots")) {
          return {
            ok: true,
            json: async () => ({ job_id: "job_4", slots: state.slots }),
          } as Response;
        }

        if (url.includes("/api/jobs/job_4")) {
          return {
            ok: true,
            json: async () => state.job,
          } as Response;
        }

        return {
          ok: false,
          status: 404,
          json: async () => ({ error: "not found", error_code: "RESOURCE_NOT_FOUND" }),
        } as Response;
      }),
    );

    render(
      <MemoryRouter initialEntries={["/jobs/job_4"]}>
        <Routes>
          <Route path="/jobs/:jobId" element={<JobPage />} />
        </Routes>
      </MemoryRouter>,
    );

    await screen.findByText("Slot ID: slot_4");
    fireEvent.change(screen.getByLabelText("Generated clip path"), {
      target: { value: "/tmp/example2/generated.mp4" },
    });
    fireEvent.click(screen.getByRole("button", { name: "Import generated clip" }));

    await screen.findByText("manual generated clip imported");
    expect(screen.getByText("/tmp/example2/generated.mp4")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Render preview" })).toBeInTheDocument();
  });
});
