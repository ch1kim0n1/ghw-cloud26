import { useState } from "react";
import { ApiError } from "../types/Api";
import type { Job } from "../types/Job";
import { getJob } from "../services/jobsApi";
import { usePollingRequest } from "./usePollingRequest";

interface UseJobOptions {
  poll?: boolean;
  pollIntervalMs?: number;
}

export function useJob(jobId?: string, options: UseJobOptions = {}) {
  const [job, setJob] = useState<Job | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  const { refresh } = usePollingRequest(
    async (showLoading, isCancelled) => {
      if (!jobId) {
        return;
      }
      if (showLoading) {
        setLoading(true);
      }

      try {
        const response = await getJob(jobId);
        if (isCancelled()) {
          return;
        }
        setJob(response);
        setError(null);
      } catch (reason: unknown) {
        if (isCancelled()) {
          return;
        }
        if (reason instanceof ApiError) {
          setError(reason.message);
        } else {
          setError("Unable to load job.");
        }
      } finally {
        if (!isCancelled()) {
          setLoading(false);
        }
      }
    },
    [jobId],
    {
      enabled: Boolean(jobId),
      poll: options.poll,
      pollIntervalMs: options.pollIntervalMs,
    },
  );

  return {
    job,
    error,
    loading,
    refresh,
  };
}
