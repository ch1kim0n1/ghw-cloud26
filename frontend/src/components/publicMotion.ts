import type { Transition, Variants } from "framer-motion";

export const publicRoutePaths = new Set(["/", "/about", "/gallery", "/studio", "/results", "/upload", "/website-ads"]);

export const publicEase: [number, number, number, number] = [0.22, 1, 0.36, 1];
export const publicExitEase: [number, number, number, number] = [0.4, 0, 1, 1];

export const publicTransition = {
  duration: 0.52,
  ease: publicEase,
} satisfies Transition;

export const publicQuickTransition = {
  duration: 0.28,
  ease: publicEase,
} satisfies Transition;

export const publicSwapTransition = {
  duration: 0.34,
  ease: publicEase,
} satisfies Transition;

export const publicLayoutTransition = {
  type: "spring",
  stiffness: 320,
  damping: 28,
  mass: 0.9,
} satisfies Transition;

export const publicStagger = 0.1;
export const publicViewport = { once: true, amount: 0.18 } as const;

export function buildEnterVariants(distance = 24): Variants {
  return {
    hidden: { opacity: 0, y: distance },
    show: {
      opacity: 1,
      y: 0,
      transition: publicTransition,
    },
    exit: {
      opacity: 0,
      y: -Math.max(10, Math.round(distance * 0.5)),
      transition: { duration: 0.22, ease: publicExitEase },
    },
  };
}

export const pageShellVariants: Variants = {
  hidden: { opacity: 0, y: 22 },
  show: { opacity: 1, y: 0, transition: publicTransition },
  exit: { opacity: 0, y: -14, transition: { duration: 0.22, ease: publicExitEase } },
};

export const staggerContainerVariants: Variants = {
  hidden: {},
  show: {
    transition: {
      staggerChildren: publicStagger,
      delayChildren: 0.04,
    },
  },
};

export const staggerItemVariants = buildEnterVariants(18);

export const contentSwapVariants = buildEnterVariants(18);
