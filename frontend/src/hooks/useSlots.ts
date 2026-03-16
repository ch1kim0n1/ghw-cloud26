import { useState } from "react";
import { ApiError } from "../types/Api";
import type { Slot } from "../types/Slot";
import { listSlots } from "../services/slotsApi";
import { usePollingRequest } from "./usePollingRequest";

interface UseSlotsOptions {
  poll?: boolean;
  pollIntervalMs?: number;
}

export function useSlots(jobId?: string, options: UseSlotsOptions = {}) {
  const [slots, setSlots] = useState<Slot[]>([]);
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
        const response = await listSlots(jobId);
        if (isCancelled()) {
          return;
        }
        setSlots(response.slots ?? []);
        setError(null);
      } catch (reason: unknown) {
        if (isCancelled()) {
          return;
        }
        if (reason instanceof ApiError) {
          setError(reason.message);
        } else {
          setError("Unable to load slots.");
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
    slots,
    error,
    loading,
    refresh,
  };
}
