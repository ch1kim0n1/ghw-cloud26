import { useEffect, useState } from "react";
import { ApiError } from "../types/Api";
import type { Job } from "../types/Job";
import { getJob } from "../services/jobsApi";

export function useJob(jobId?: string) {
  const [job, setJob] = useState<Job | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if (!jobId) {
      return;
    }

    setLoading(true);
    void getJob(jobId)
      .then((response) => {
        setJob(response);
        setError(null);
      })
      .catch((reason: unknown) => {
        if (reason instanceof ApiError) {
          setError(reason.message);
          return;
        }
        setError("Unable to load job.");
      })
      .finally(() => setLoading(false));
  }, [jobId]);

  return { job, error, loading };
}
