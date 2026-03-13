import { render, screen, waitFor } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import App from "./App";

describe("App", () => {
  beforeEach(() => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue({
        ok: true,
        json: async () => ({
          status: "healthy",
          timestamp: "2026-03-13T00:00:00Z",
          version: "0.1.0-mvp",
          provider_profile: "azure",
        }),
      }),
    );
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("renders the products page route", async () => {
    render(
      <MemoryRouter initialEntries={["/products"]}>
        <App />
      </MemoryRouter>,
    );

    expect(screen.getByText("Product Catalog")).toBeInTheDocument();
    await waitFor(() => {
      expect(screen.getByText("healthy")).toBeInTheDocument();
    });
    expect(screen.getByText("Provider azure")).toBeInTheDocument();
    expect(screen.queryByRole("link", { name: "Preview" })).not.toBeInTheDocument();
  });

  it("renders the campaign page route", () => {
    render(
      <MemoryRouter initialEntries={["/campaigns/new"]}>
        <App />
      </MemoryRouter>,
    );

    expect(screen.getByText("Campaign Intake")).toBeInTheDocument();
  });

  it("renders the job page route", () => {
    render(
      <MemoryRouter initialEntries={["/jobs/test-job"]}>
        <App />
      </MemoryRouter>,
    );

    expect(screen.getByText(/Job dashboard test-job/)).toBeInTheDocument();
  });

  it("renders the preview page route", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(async (input: RequestInfo | URL) => {
        const url = String(input);
        if (url.includes("/api/health")) {
          return {
            ok: true,
            json: async () => ({
              status: "healthy",
              timestamp: "2026-03-13T00:00:00Z",
              version: "0.1.0-mvp",
              provider_profile: "azure",
            }),
          } as Response;
        }
        return {
          ok: false,
          status: 404,
          json: async () => ({
            error: "preview not found",
            error_code: "RESOURCE_NOT_FOUND",
          }),
        } as Response;
      }),
    );

    render(
      <MemoryRouter initialEntries={["/jobs/test-job/preview"]}>
        <App />
      </MemoryRouter>,
    );

    await waitFor(() => {
      expect(screen.getByText(/Preview status test-job/)).toBeInTheDocument();
    });
    await waitFor(() => {
      expect(screen.getByText("No preview has been started yet.")).toBeInTheDocument();
    });
  });

  it("shows an error state when health fails", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue({
        ok: false,
        status: 503,
        json: async () => ({
          error: "backend unavailable",
          error_code: "SERVICE_UNAVAILABLE",
        }),
      }),
    );

    render(
      <MemoryRouter initialEntries={["/products"]}>
        <App />
      </MemoryRouter>,
    );

    await waitFor(() => {
      expect(screen.getByText("Connection failed")).toBeInTheDocument();
    });
  });
});
