import Particles, { initParticlesEngine } from "@tsparticles/react";
import { loadSlim } from "@tsparticles/slim";
import { useEffect, useState } from "react";

export function AmbientParticles() {
  const [ready, setReady] = useState(false);

  useEffect(() => {
    if (import.meta.env.MODE === "test") {
      return;
    }

    let mounted = true;

    initParticlesEngine(async (engine) => {
      await loadSlim(engine);
    }).then(() => {
      if (mounted) {
        setReady(true);
      }
    });

    return () => {
      mounted = false;
    };
  }, []);

  if (!ready || import.meta.env.MODE === "test") {
    return null;
  }

  return (
    <Particles
      className="ambient-particles"
      options={{
        background: {
          color: {
            value: "transparent",
          },
        },
        fullScreen: {
          enable: false,
        },
        fpsLimit: 60,
        particles: {
          number: {
            value: 26,
          },
          color: {
            value: ["#ff82b8", "#ffd8ec", "#fff7c0", "#ffffff"],
          },
          move: {
            direction: "top",
            enable: true,
            outModes: {
              default: "out",
            },
            speed: 0.6,
          },
          opacity: {
            value: {
              min: 0.2,
              max: 0.65,
            },
          },
          rotate: {
            value: {
              min: 0,
              max: 360,
            },
            animation: {
              enable: true,
              speed: 8,
            },
          },
          shape: {
            type: ["square", "star"],
          },
          size: {
            value: {
              min: 3,
              max: 8,
            },
          },
        },
        detectRetina: true,
      }}
    />
  );
}
