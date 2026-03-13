import { useEffect, useState } from "react";
import { ApiError } from "../types/Api";
import type { Preview } from "../types/Preview";
import { getPreview } from "../services/previewApi";

export function usePreview(jobId?: string) {
  const [preview, setPreview] = useState<Preview | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if (!jobId) {
      return;
    }

    setLoading(true);
    void getPreview(jobId)
      .then((response) => {
        setPreview(response);
        setError(null);
      })
      .catch((reason: unknown) => {
        if (reason instanceof ApiError) {
          setError(reason.message);
          return;
        }
        setError("Unable to load preview placeholder.");
      })
      .finally(() => setLoading(false));
  }, [jobId]);

  return { preview, error, loading };
}
