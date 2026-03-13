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

    expect(screen.getByText(/Job placeholder test-job/)).toBeInTheDocument();
  });

  it("renders the preview page route", () => {
    render(
      <MemoryRouter initialEntries={["/preview/test-job"]}>
        <App />
      </MemoryRouter>,
    );

    expect(screen.getByText(/Preview scaffold test-job/)).toBeInTheDocument();
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
