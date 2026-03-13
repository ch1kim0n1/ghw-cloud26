import { useEffect, useState } from "react";
import { ApiError } from "../types/Api";
import type { JobLog } from "../types/Job";
import { getJobLogs } from "../services/jobsApi";

interface UseJobLogsOptions {
  poll?: boolean;
  pollIntervalMs?: number;
}

export function useJobLogs(jobId?: string, options: UseJobLogsOptions = {}) {
  const [logs, setLogs] = useState<JobLog[]>([]);
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

      return getJobLogs(jobId)
        .then((response) => {
          if (cancelled) {
            return;
          }
          setLogs(response.logs ?? []);
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
          setError("Unable to load job logs.");
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
    logs,
    error,
    loading,
    refresh: () => setRefreshKey((value) => value + 1),
  };
}
