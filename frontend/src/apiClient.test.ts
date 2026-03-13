import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { buildApiUrl, request } from "./services/apiClient";
import { listProducts } from "./services/productsApi";
import { ApiError } from "./types/Api";

describe("apiClient", () => {
  beforeEach(() => {
    vi.stubGlobal("fetch", vi.fn());
  });

  afterEach(() => {
    vi.unstubAllGlobals();
    vi.unstubAllEnvs();
  });

  it("uses the configured base url", () => {
    vi.stubEnv("VITE_API_BASE_URL", "http://localhost:9000");
    expect(buildApiUrl("/api/health")).toBe("http://localhost:9000/api/health");
  });

  it("normalizes placeholder api errors", async () => {
    vi.mocked(fetch).mockResolvedValue({
      ok: false,
      status: 501,
      json: async () => ({
        error: "endpoint not implemented in phase 0",
        error_code: "NOT_IMPLEMENTED",
        details: {
          path: "/api/products",
        },
      }),
    } as Response);

    await expect(listProducts()).rejects.toMatchObject({
      status: 501,
      code: "NOT_IMPLEMENTED",
    } satisfies Partial<ApiError>);
  });

  it("returns typed json payloads", async () => {
    vi.mocked(fetch).mockResolvedValue({
      ok: true,
      json: async () => ({
        status: "healthy",
      }),
    } as Response);

    const payload = await request<{ status: string }>("/api/health");
    expect(payload.status).toBe("healthy");
  });
});
