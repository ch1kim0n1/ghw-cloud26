import { useEffect, useState } from "react";
import { ApiError, type HealthResponse } from "../types/Api";
import { getHealth } from "../services/healthApi";

interface UseHealthOptions {
  poll?: boolean;
  pollIntervalMs?: number;
}

export function useHealth(options: UseHealthOptions = {}) {
  const [health, setHealth] = useState<HealthResponse | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const [refreshKey, setRefreshKey] = useState(0);

  useEffect(() => {
    let cancelled = false;

    const load = (showLoading: boolean) => {
      if (showLoading) {
        setLoading(true);
      }

      return getHealth()
        .then((response) => {
          if (cancelled) {
            return;
          }
          setHealth(response);
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
          setError("Unable to load health status.");
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
      }, options.pollIntervalMs ?? 5000);
    }

    return () => {
      cancelled = true;
      if (intervalId !== undefined) {
        window.clearInterval(intervalId);
      }
    };
  }, [options.poll, options.pollIntervalMs, refreshKey]);

  return {
    health,
    error,
    loading,
    refresh: () => setRefreshKey((value) => value + 1),
  };
}
