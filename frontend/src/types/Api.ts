export interface HealthResponse {
  status: string;
  timestamp: string;
  version: string;
  provider_profile: string;
}

export interface ApiErrorPayload {
  error: string;
  error_code?: string;
  http_status?: number;
  details?: Record<string, unknown>;
  timestamp?: string;
}

export class ApiError extends Error {
  status: number;
  code?: string;
  details?: Record<string, unknown>;

  constructor(message: string, status: number, code?: string, details?: Record<string, unknown>) {
    super(message);
    this.name = "ApiError";
    this.status = status;
    this.code = code;
    this.details = details;
  }
}
