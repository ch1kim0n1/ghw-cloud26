import Lenis from "lenis";
import { useEffect } from "react";
import { useLocation } from "react-router-dom";
import { publicRoutePaths } from "./publicMotion";

export function PublicMotionProvider() {
  const location = useLocation();

  useEffect(() => {
    if (import.meta.env.MODE === "test") {
      return;
    }

    const isPublicRoute = publicRoutePaths.has(location.pathname);
    const prefersReducedMotion = window.matchMedia("(prefers-reduced-motion: reduce)").matches;

    if (!isPublicRoute || prefersReducedMotion) {
      return;
    }

    const lenis = new Lenis({
      duration: 0.92,
      smoothWheel: true,
      touchMultiplier: 1.02,
    });

    let frameId = 0;

    const raf = (time: number) => {
      lenis.raf(time);
      frameId = window.requestAnimationFrame(raf);
    };

    frameId = window.requestAnimationFrame(raf);

    return () => {
      window.cancelAnimationFrame(frameId);
      lenis.destroy();
    };
  }, [location.pathname]);

  return null;
}
