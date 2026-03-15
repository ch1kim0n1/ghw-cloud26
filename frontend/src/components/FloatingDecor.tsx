import { animated, useSprings } from "@react-spring/web";
import { useReducedMotion } from "framer-motion";

const decorAssets = {
  bow: "/decor/voxel-bow.svg",
  cloud: "/decor/voxel-cloud.svg",
  flower: "/decor/voxel-flower.svg",
  heart: "/decor/voxel-heart.svg",
  star: "/decor/voxel-star.svg",
} as const;

type DecorAssetId = keyof typeof decorAssets;

type FloatingDecorProps = {
  ids: readonly string[];
  variant?: "hero" | "upload" | "about";
};

export function FloatingDecor({ ids, variant = "hero" }: FloatingDecorProps) {
  const safeIds = ids.filter((id): id is DecorAssetId => id in decorAssets);
  const reducedMotion = useReducedMotion();
  const springs = useSprings(
    safeIds.length,
    safeIds.map((id, index) => ({
      from: {
        opacity: 0,
        transform: `translate3d(0, 14px, 0) rotate(${index % 2 === 0 ? -3 : 3}deg) scale(0.94)`,
      },
      to: {
        opacity: 1,
        transform: reducedMotion
          ? "translate3d(0, 0, 0) rotate(0deg) scale(1)"
          : `translate3d(0, ${index % 2 === 0 ? -5 : 4}px, 0) rotate(${index % 2 === 0 ? 2 : -2}deg) scale(1)`,
      },
      loop: reducedMotion
        ? false
        : {
            reverse: true,
          },
      delay: 180 + index * 140,
      config: {
        mass: 1.55,
        tension: 88 + index * 4,
        friction: 20,
      },
    })),
  );

  return (
    <div className={`floating-decor floating-decor--${variant}`} aria-hidden="true">
      {springs.map((style, index) => (
        <animated.img
          className={`floating-decor__item floating-decor__item--${index + 1}`}
          key={`${safeIds[index]}-${index}`}
          src={decorAssets[safeIds[index]]}
          alt=""
          style={style}
        />
      ))}
    </div>
  );
}
