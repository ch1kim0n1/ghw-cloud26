import { ApiError, type ApiErrorPayload } from "../types/Api";

const DEFAULT_API_BASE_URL = "http://localhost:8080";

export function getApiBaseUrl(): string {
  return import.meta.env.VITE_API_BASE_URL ?? DEFAULT_API_BASE_URL;
}

export function buildApiUrl(path: string): string {
  return new URL(path, getApiBaseUrl()).toString();
}

export async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(buildApiUrl(path), {
    headers: {
      Accept: "application/json",
      ...(init?.headers ?? {}),
    },
    ...init,
  });

  if (!response.ok) {
    let payload: ApiErrorPayload | undefined;
    try {
      payload = (await response.json()) as ApiErrorPayload;
    } catch {
      payload = undefined;
    }

    throw new ApiError(
      payload?.error ?? `request failed with status ${response.status}`,
      response.status,
      payload?.error_code,
      payload?.details,
    );
  }

  return (await response.json()) as T;
}
