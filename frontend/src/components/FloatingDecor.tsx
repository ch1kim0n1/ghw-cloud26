import { animated, useSprings } from "@react-spring/web";

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
  const springs = useSprings(
    safeIds.length,
    safeIds.map((id, index) => ({
      from: {
        opacity: 0,
        transform: `translate3d(0, 16px, 0) rotate(${index % 2 === 0 ? -4 : 4}deg) scale(0.92)`,
      },
      to: {
        opacity: 1,
        transform: `translate3d(0, ${index % 2 === 0 ? -8 : 6}px, 0) rotate(${index % 2 === 0 ? 3 : -3}deg) scale(1)`,
      },
      loop: {
        reverse: true,
      },
      delay: 180 + index * 140,
      config: {
        mass: 1.4,
        tension: 105 + index * 6,
        friction: 18,
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
