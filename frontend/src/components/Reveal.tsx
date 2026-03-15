import { motion, useReducedMotion } from "framer-motion";
import type { PropsWithChildren, ReactNode } from "react";
import { buildEnterVariants, publicStagger, publicViewport, staggerContainerVariants, staggerItemVariants } from "./publicMotion";

type RevealProps = PropsWithChildren<{
  className?: string;
  delay?: number;
  as?: "div" | "section" | "article";
  distance?: number;
}>;

export function Reveal({ children, className, delay = 0, as = "div", distance = 24 }: RevealProps) {
  const reducedMotion = useReducedMotion();
  const Component = as === "section" ? motion.section : as === "article" ? motion.article : motion.div;

  if (reducedMotion) {
    return <Component className={className}>{children}</Component>;
  }

  return (
    <Component
      className={className}
      initial="hidden"
      whileInView="show"
      viewport={publicViewport}
      variants={buildEnterVariants(distance)}
      transition={{ delay }}
    >
      {children}
    </Component>
  );
}

type StaggerListProps = {
  className?: string;
  children: ReactNode;
};

export function StaggerList({ className, children }: StaggerListProps) {
  const reducedMotion = useReducedMotion();

  if (reducedMotion) {
    return <div className={className}>{children}</div>;
  }

  return (
    <motion.div
      className={className}
      initial="hidden"
      whileInView="show"
      viewport={publicViewport}
      variants={staggerContainerVariants}
      transition={{ staggerChildren: publicStagger }}
    >
      {children}
    </motion.div>
  );
}

type StaggerItemProps = PropsWithChildren<{
  className?: string;
  as?: "div" | "article";
}>;

export function StaggerItem({ children, className, as = "div" }: StaggerItemProps) {
  const reducedMotion = useReducedMotion();
  const Component = as === "article" ? motion.article : motion.div;

  if (reducedMotion) {
    return <Component className={className}>{children}</Component>;
  }

  return (
    <Component
      className={className}
      variants={staggerItemVariants}
    >
      {children}
    </Component>
  );
}
