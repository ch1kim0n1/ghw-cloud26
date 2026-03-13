import { useEffect, useState } from "react";
import { ApiError } from "../types/Api";
import type { Slot } from "../types/Slot";
import { listSlots } from "../services/slotsApi";

export function useSlots(jobId?: string) {
  const [slots, setSlots] = useState<Slot[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if (!jobId) {
      return;
    }

    setLoading(true);
    void listSlots(jobId)
      .then((response) => {
        setSlots(response.slots);
        setError(null);
      })
      .catch((reason: unknown) => {
        if (reason instanceof ApiError) {
          setError(reason.message);
          return;
        }
        setError("Unable to load slot placeholders.");
      })
      .finally(() => setLoading(false));
  }, [jobId]);

  return { slots, error, loading };
}
