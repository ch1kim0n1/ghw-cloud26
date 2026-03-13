import { request } from "./apiClient";
import type { HealthResponse } from "../types/Api";

export function getHealth(): Promise<HealthResponse> {
  return request<HealthResponse>("/api/health");
}
