import { useState } from "react";
import { ApiError } from "../types/Api";
import type { Preview } from "../types/Preview";
import { getPreview } from "../services/previewApi";
import { usePollingRequest } from "./usePollingRequest";

interface UsePreviewOptions {
  poll?: boolean;
  pollIntervalMs?: number;
}

export function usePreview(jobId?: string, options: UsePreviewOptions = {}) {
  const [preview, setPreview] = useState<Preview | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const shouldPoll = options.poll ?? (preview?.status === "pending" || preview?.status === "stitching");

  const { refresh } = usePollingRequest(
    async (showLoading, isCancelled) => {
      if (!jobId) {
        return;
      }
      if (showLoading) {
        setLoading(true);
      }

      try {
        const response = await getPreview(jobId);
        if (isCancelled()) {
          return;
        }
        setPreview(response);
        setError(null);
      } catch (reason: unknown) {
        if (isCancelled()) {
          return;
        }
        if (reason instanceof ApiError && reason.status === 404) {
          setPreview(null);
          setError(null);
          return;
        }
        if (reason instanceof ApiError) {
          setError(reason.message);
        } else {
          setError("Unable to load preview.");
        }
      } finally {
        if (!isCancelled()) {
          setLoading(false);
        }
      }
    },
    [jobId, shouldPoll],
    {
      enabled: Boolean(jobId),
      poll: shouldPoll,
      pollIntervalMs: options.pollIntervalMs,
    },
  );

  return {
    preview,
    error,
    loading,
    refresh,
  };
}
