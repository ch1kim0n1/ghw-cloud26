import { useState } from "react";
import { ApiError, type HealthResponse } from "../types/Api";
import { getHealth } from "../services/healthApi";
import { usePollingRequest } from "./usePollingRequest";

interface UseHealthOptions {
  poll?: boolean;
  pollIntervalMs?: number;
}

export function useHealth(options: UseHealthOptions = {}) {
  const [health, setHealth] = useState<HealthResponse | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  const { refresh } = usePollingRequest(
    async (showLoading, isCancelled) => {
      if (showLoading) {
        setLoading(true);
      }

      try {
        const response = await getHealth();
        if (isCancelled()) {
          return;
        }
        setHealth(response);
        setError(null);
      } catch (reason: unknown) {
        if (isCancelled()) {
          return;
        }
        if (reason instanceof ApiError) {
          setError(reason.message);
        } else {
          setError("Unable to load health status.");
        }
      } finally {
        if (!isCancelled()) {
          setLoading(false);
        }
      }
    },
    [],
    {
      poll: options.poll,
      pollIntervalMs: options.pollIntervalMs ?? 5000,
    },
  );

  return {
    health,
    error,
    loading,
    refresh,
  };
}
