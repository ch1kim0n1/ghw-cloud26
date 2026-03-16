import { useState } from "react";
import { ApiError } from "../types/Api";
import type { JobLog } from "../types/Job";
import { getJobLogs } from "../services/jobsApi";
import { usePollingRequest } from "./usePollingRequest";

interface UseJobLogsOptions {
  poll?: boolean;
  pollIntervalMs?: number;
}

export function useJobLogs(jobId?: string, options: UseJobLogsOptions = {}) {
  const [logs, setLogs] = useState<JobLog[]>([]);
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
        const response = await getJobLogs(jobId);
        if (isCancelled()) {
          return;
        }
        setLogs(response.logs ?? []);
        setError(null);
      } catch (reason: unknown) {
        if (isCancelled()) {
          return;
        }
        if (reason instanceof ApiError) {
          setError(reason.message);
        } else {
          setError("Unable to load job logs.");
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
    logs,
    error,
    loading,
    refresh,
  };
}
