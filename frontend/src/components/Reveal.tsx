import { motion, useReducedMotion } from "framer-motion";
import type { PropsWithChildren, ReactNode } from "react";

type RevealProps = PropsWithChildren<{
  className?: string;
  delay?: number;
  as?: "div" | "section" | "article";
}>;

export function Reveal({ children, className, delay = 0, as = "div" }: RevealProps) {
  const reducedMotion = useReducedMotion();
  const Component = as === "section" ? motion.section : as === "article" ? motion.article : motion.div;

  if (reducedMotion) {
    return <Component className={className}>{children}</Component>;
  }

  return (
    <Component
      className={className}
      initial={{ opacity: 0, y: 24 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.65, ease: [0.22, 1, 0.36, 1], delay }}
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
      animate="show"
      variants={{
        hidden: {},
        show: {
          transition: {
            staggerChildren: 0.1,
          },
        },
      }}
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
      variants={{
        hidden: { opacity: 0, y: 18 },
        show: { opacity: 1, y: 0 },
      }}
      transition={{ duration: 0.55, ease: [0.22, 1, 0.36, 1] }}
    >
      {children}
    </Component>
  );
}
