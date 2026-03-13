import { useEffect, useState } from "react";
import { ApiError } from "../types/Api";
import type { Preview } from "../types/Preview";
import { getPreview } from "../services/previewApi";

interface UsePreviewOptions {
  poll?: boolean;
  pollIntervalMs?: number;
}

export function usePreview(jobId?: string, options: UsePreviewOptions = {}) {
  const [preview, setPreview] = useState<Preview | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const [refreshKey, setRefreshKey] = useState(0);
  const shouldPoll = options.poll ?? (preview?.status === "pending" || preview?.status === "stitching");

  useEffect(() => {
    if (!jobId) {
      return;
    }

    let cancelled = false;

    const load = (showLoading: boolean) => {
      if (showLoading) {
        setLoading(true);
      }

      return getPreview(jobId)
        .then((response) => {
          if (cancelled) {
            return;
          }
          setPreview(response);
          setError(null);
        })
        .catch((reason: unknown) => {
          if (cancelled) {
            return;
          }
          if (reason instanceof ApiError && reason.status === 404) {
            setPreview(null);
            setError(null);
            return;
          }
          if (reason instanceof ApiError) {
            setError(reason.message);
            return;
          }
          setError("Unable to load preview.");
        })
        .finally(() => {
          if (!cancelled) {
            setLoading(false);
          }
        });
    };

    void load(true);

    let intervalId: number | undefined;
    if (shouldPoll) {
      intervalId = window.setInterval(() => {
        void load(false);
      }, options.pollIntervalMs ?? 2000);
    }

    return () => {
      cancelled = true;
      if (intervalId !== undefined) {
        window.clearInterval(intervalId);
      }
    };
  }, [jobId, options.poll, options.pollIntervalMs, refreshKey, shouldPoll]);

  return {
    preview,
    error,
    loading,
    refresh: () => setRefreshKey((value) => value + 1),
  };
}
