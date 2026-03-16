import { DependencyList, useEffect, useState } from "react";

interface UsePollingRequestOptions {
  enabled?: boolean;
  poll?: boolean;
  pollIntervalMs?: number;
}

export function usePollingRequest(
  load: (showLoading: boolean, isCancelled: () => boolean) => Promise<void>,
  dependencies: DependencyList,
  options: UsePollingRequestOptions = {},
) {
  const [refreshKey, setRefreshKey] = useState(0);

  useEffect(() => {
    if (options.enabled === false) {
      return;
    }

    let cancelled = false;
    const isCancelled = () => cancelled;

    void load(true, isCancelled);

    let intervalId: number | undefined;
    if (options.poll) {
      intervalId = window.setInterval(() => {
        void load(false, isCancelled);
      }, options.pollIntervalMs ?? 2000);
    }

    return () => {
      cancelled = true;
      if (intervalId !== undefined) {
        window.clearInterval(intervalId);
      }
    };
  }, [...dependencies, options.enabled, options.poll, options.pollIntervalMs, refreshKey]);

  return {
    refresh: () => setRefreshKey((value) => value + 1),
  };
}
