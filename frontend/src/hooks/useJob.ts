import { useEffect, useState } from "react";
import { ApiError } from "../types/Api";
import type { Job } from "../types/Job";
import { getJob } from "../services/jobsApi";

interface UseJobOptions {
  poll?: boolean;
  pollIntervalMs?: number;
}

export function useJob(jobId?: string, options: UseJobOptions = {}) {
  const [job, setJob] = useState<Job | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const [refreshKey, setRefreshKey] = useState(0);

  useEffect(() => {
    if (!jobId) {
      return;
    }

    let cancelled = false;

    const load = (showLoading: boolean) => {
      if (showLoading) {
        setLoading(true);
      }

      return getJob(jobId)
      .then((response) => {
        if (cancelled) {
          return;
        }
        setJob(response);
        setError(null);
      })
      .catch((reason: unknown) => {
        if (cancelled) {
          return;
        }
        if (reason instanceof ApiError) {
          setError(reason.message);
          return;
        }
        setError("Unable to load job.");
      })
      .finally(() => {
        if (!cancelled) {
          setLoading(false);
        }
      });
    };

    void load(true);

    let intervalId: number | undefined;
    if (options.poll) {
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
  }, [jobId, options.poll, options.pollIntervalMs, refreshKey]);

  return {
    job,
    error,
    loading,
    refresh: () => setRefreshKey((value) => value + 1),
  };
}
