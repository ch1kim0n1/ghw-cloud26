import { cleanup, render, screen, waitFor } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import App from "./App";

describe("App", () => {
  beforeEach(() => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue({
      ok: true,
      json: async () => ({ products: [] }),
    }));
  });

  afterEach(() => {
    cleanup();
    vi.unstubAllGlobals();
  });

  it("renders the simplified showcase route with only two visible tabs", () => {
    render(
      <MemoryRouter initialEntries={["/"]}>
        <App />
      </MemoryRouter>,
    );

    expect(screen.getByRole("link", { name: "Showcase" })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Upload" })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "About us" })).toBeInTheDocument();
    expect(screen.queryByRole("link", { name: "Results" })).not.toBeInTheDocument();
    expect(screen.queryByRole("link", { name: "Studio dashboard" })).not.toBeInTheDocument();
    expect(screen.getByText("Ad insertion, but make it cute, seamless, and actually watchable.")).toBeInTheDocument();
    expect(screen.getByText("Outdoor reveal with a late-scene handoff")).toBeInTheDocument();
    expect(screen.getByText("Talking-head scene with a seamless branded bridge")).toBeInTheDocument();
    expect(screen.getByText("Streamer close-up with an early energy-drink insert")).toBeInTheDocument();
  });

  it("renders the upload route with the simplified fields", () => {
    render(
      <MemoryRouter initialEntries={["/upload"]}>
        <App />
      </MemoryRouter>,
    );

    expect(screen.getByLabelText("Campaign name")).toBeInTheDocument();
    expect(screen.getByLabelText("Brand / Product name")).toBeInTheDocument();
    expect(screen.getByLabelText("Source video")).toBeInTheDocument();
    expect(screen.queryByText("Product source")).not.toBeInTheDocument();
    expect(screen.queryByLabelText("Source URL")).not.toBeInTheDocument();
  });

  it("keeps the hidden results route accessible", () => {
    render(
      <MemoryRouter initialEntries={["/results"]}>
        <App />
      </MemoryRouter>,
    );

    expect(screen.getByText("Ad insertion, but make it cute, seamless, and actually watchable.")).toBeInTheDocument();
  });

  it("renders the about route with developer placeholders", () => {
    render(
      <MemoryRouter initialEntries={["/about"]}>
        <App />
      </MemoryRouter>,
    );

    expect(screen.getByText("Two builders, one cute little ad-insertion experiment.")).toBeInTheDocument();
    expect(screen.getByText("Developer One")).toBeInTheDocument();
    expect(screen.getByText("Developer Two")).toBeInTheDocument();
  });

  it("keeps the hidden products route accessible", async () => {
    render(
      <MemoryRouter initialEntries={["/products"]}>
        <App />
      </MemoryRouter>,
    );

    await waitFor(() => {
      expect(screen.getByText("Product Catalog")).toBeInTheDocument();
    });
    expect(screen.queryByRole("link", { name: "Studio dashboard" })).not.toBeInTheDocument();
  });

  it("keeps the hidden campaign page accessible", () => {
    render(
      <MemoryRouter initialEntries={["/campaigns/new"]}>
        <App />
      </MemoryRouter>,
    );

    expect(screen.getByText("Campaign Intake")).toBeInTheDocument();
  });

  it("keeps the hidden job page accessible", () => {
    render(
      <MemoryRouter initialEntries={["/jobs/test-job"]}>
        <App />
      </MemoryRouter>,
    );

    expect(screen.getByText(/Job dashboard test-job/)).toBeInTheDocument();
  });

  it("keeps the hidden preview page accessible", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(async (input: RequestInfo | URL) => {
        const url = String(input);
        if (url.includes("/api/jobs/test-job/preview")) {
          return {
            ok: false,
            status: 404,
            json: async () => ({
              error: "preview not found",
              error_code: "RESOURCE_NOT_FOUND",
            }),
          } as Response;
        }

        return {
          ok: true,
          json: async () => ({ products: [] }),
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
  });
});
