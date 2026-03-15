import { cleanup, fireEvent, render, screen, waitFor, within } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import App from "./App";

describe("App", () => {
  beforeEach(() => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue({
        ok: true,
        json: async () => ({ products: [] }),
      }),
    );
  });

  afterEach(() => {
    cleanup();
    vi.unstubAllGlobals();
  });

  it("renders the landing route with the voxel nav, hero, proof rail, and upload cta", () => {
    render(
      <MemoryRouter initialEntries={["/"]}>
        <App />
      </MemoryRouter>,
    );

    expect(screen.getByRole("link", { name: "Home" })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Gallery" })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Upload" })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "About" })).toBeInTheDocument();
    expect(screen.queryByRole("link", { name: "Results" })).not.toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "CAFAI turns product insertion into a scene-aware, watchable cut." })).toBeInTheDocument();
    expect(screen.getByText("One featured scene. Four receipts right under it.")).toBeInTheDocument();
    expect(screen.getByRole("tab", { name: /Final stitched idol cut/i })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Start an upload" })).toBeInTheDocument();
  });

  it("switches the featured hero example from the demo selector", () => {
    render(
      <MemoryRouter initialEntries={["/"]}>
        <App />
      </MemoryRouter>,
    );

    const demoSelector = screen.getByRole("tablist", { name: "Featured demo selector" });
    fireEvent.click(within(demoSelector).getByRole("tab", { name: /Bike Bloom/i }));

    expect(screen.getAllByRole("heading", { name: "Bike Bloom Reveal" })[0]).toBeInTheDocument();
    expect(screen.getByText("41.708s -> 43.377s")).toBeInTheDocument();
  });

  it("renders the upload route with the custom dropzone", () => {
    render(
      <MemoryRouter initialEntries={["/upload"]}>
        <App />
      </MemoryRouter>,
    );

    expect(screen.getByLabelText("Campaign name")).toBeInTheDocument();
    expect(screen.getByLabelText("Brand / Product name")).toBeInTheDocument();
    expect(screen.getByLabelText("Source video")).toBeInTheDocument();
    expect(screen.getByText("Drag an MP4 here or tap to browse.")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Upload a product" })).toBeInTheDocument();
    expect(screen.queryByText("Product source")).not.toBeInTheDocument();
    expect(screen.queryByLabelText("Source URL")).not.toBeInTheDocument();
  });

  it("keeps the hidden results route accessible", () => {
    render(
      <MemoryRouter initialEntries={["/results"]}>
        <App />
      </MemoryRouter>,
    );

    expect(screen.getByText("Gallery of processed videos")).toBeInTheDocument();
    expect(screen.getByRole("tablist", { name: "Showcase examples" })).toBeInTheDocument();
  });

  it("renders the about route with the upgraded founder cards", () => {
    render(
      <MemoryRouter initialEntries={["/about"]}>
        <App />
      </MemoryRouter>,
    );

    expect(screen.getByText("The two developers behind the CAFAI demo.")).toBeInTheDocument();
    expect(screen.getByText("Vlad")).toBeInTheDocument();
    expect(screen.getByText("Monika Jaqeli")).toBeInTheDocument();
    expect(screen.getByAltText("Vlad profile meme")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Vlad GitHub profile" })).toHaveAttribute("href", "https://github.com/ch1kim0n1");
    expect(screen.getByRole("link", { name: "Monika Jaqeli GitHub profile" })).toHaveAttribute(
      "href",
      "https://github.com/SuperLepeshka",
    );
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
