import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { JobPage } from "./pages/JobPage";

describe("Phase 2 job page", () => {
  afterEach(() => {
    cleanup();
    vi.unstubAllGlobals();
  });

  it("starts analysis, shows slots, and gates re-pick until all slots are rejected", async () => {
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

        if (url.includes("/api/jobs/job_1/slots/slot_1/reject") && init?.method === "POST") {
          state.slots = state.slots.map((slot) =>
            slot.id === "slot_1" ? { ...slot, status: "rejected" } : slot,
          );
          return {
            ok: true,
            json: async () => ({
              job_id: "job_1",
              slot_id: "slot_1",
              slot_status: "rejected",
              message: "slot rejected",
            }),
          } as Response;
        }

        if (url.includes("/api/jobs/job_1/slots/slot_2/reject") && init?.method === "POST") {
          state.slots = state.slots.map((slot) =>
            slot.id === "slot_2" ? { ...slot, status: "rejected" } : slot,
          );
          state.job = {
            ...state.job,
            metadata: {
              ...(state.job.metadata as Record<string, unknown>),
              rejected_slot_ids: ["slot_1", "slot_2"],
            },
          };
          return {
            ok: true,
            json: async () => ({
              job_id: "job_1",
              slot_id: "slot_2",
              slot_status: "rejected",
              message: "slot rejected",
            }),
          } as Response;
        }

        if (url.includes("/api/jobs/job_1/slots/re-pick") && init?.method === "POST") {
          state.job = {
            ...state.job,
            metadata: {
              ...(state.job.metadata as Record<string, unknown>),
              repick_count: 1,
              top_slot_ids: ["slot_3"],
            },
          };
          state.slots = [
            {
              id: "slot_3",
              rank: 1,
              scene_id: "scene_3",
              anchor_start_frame: 360,
              anchor_end_frame: 361,
              source_fps: 24,
              quiet_window_seconds: 3.5,
              score: 0.79,
              reasoning: "replacement candidate",
              status: "proposed",
            },
          ];
          return {
            ok: true,
            json: async () => ({
              job_id: "job_1",
              status: "analyzing",
              current_stage: "slot_selection",
              message: "re-pick requested",
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
    expect(screen.queryByText("Product Line Review")).not.toBeInTheDocument();

    const repickButton = screen.getByRole("button", { name: "Re-pick slots" });
    expect(repickButton).toBeDisabled();

    fireEvent.click(screen.getByRole("button", { name: "Start analysis" }));

    await screen.findByText("top ranked candidate");
    expect(screen.getByText("stage_started")).toBeInTheDocument();

    fireEvent.click(screen.getAllByRole("button", { name: "Reject" })[0]);
    await screen.findByText("slot rejected");
    expect(repickButton).toBeDisabled();

    fireEvent.click(screen.getAllByRole("button", { name: "Reject" })[1]);
    await waitFor(() => {
      expect(repickButton).not.toBeDisabled();
    });

    fireEvent.click(repickButton);

    await screen.findByText("replacement candidate");
    await waitFor(() => {
      expect(screen.getByText("1 slot(s)")).toBeInTheDocument();
    });
  });
});
