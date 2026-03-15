import { useEffect, useState } from "react";
import { getHealth } from "../services/healthApi";
import { ApiError, type HealthResponse } from "../types/Api";

export function HealthStatusPanel() {
  const [health, setHealth] = useState<HealthResponse | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;

    void getHealth()
      .then((response) => {
        if (!cancelled) {
          setHealth(response);
          setError(null);
        }
      })
      .catch((reason: unknown) => {
        if (!cancelled) {
          if (reason instanceof ApiError) {
            setError(reason.message);
            return;
          }
          setError("backend unavailable");
        }
      });

    return () => {
      cancelled = true;
    };
  }, []);

  if (error) {
    return (
      <section className="panel health-panel health-panel--error">
        <p className="eyebrow">System check</p>
        <h2>Connection failed</h2>
        <p>{error}</p>
      </section>
    );
  }

  if (!health) {
    return (
      <section className="panel health-panel">
        <p className="eyebrow">System check</p>
        <h2>Checking live readiness</h2>
        <p>Verifying the local control plane before the demo goes on screen.</p>
      </section>
    );
  }

  return (
    <section className="panel health-panel health-panel--ok">
      <p className="eyebrow">System check</p>
      <h2>{health.status}</h2>
      <p>Runtime {health.version}</p>
      <p>Provider {health.provider_profile}</p>
      <p className="muted">Last response: {health.timestamp}</p>
    </section>
  );
}
