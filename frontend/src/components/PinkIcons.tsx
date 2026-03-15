import { SVGProps } from "react";

type IconProps = SVGProps<SVGSVGElement>;

function baseProps(props: IconProps) {
  return {
    viewBox: "0 0 24 24",
    fill: "none",
    stroke: "currentColor",
    strokeWidth: 1.8,
    strokeLinecap: "round" as const,
    strokeLinejoin: "round" as const,
    "aria-hidden": true,
    ...props,
  };
}

export function SparkleIcon(props: IconProps) {
  return (
    <svg {...baseProps(props)}>
      <path d="M12 3.5 13.8 8l4.7 1.8-4.7 1.8L12 16l-1.8-4.4L5.5 9.8 10.2 8 12 3.5Z" />
      <path d="M18.5 15.5 19.3 17.4l1.9.8-1.9.8-.8 1.9-.8-1.9-1.9-.8 1.9-.8.8-1.9Z" />
      <path d="M5 15.2 5.6 17l1.8.6-1.8.6-.6 1.8-.6-1.8-1.8-.6 1.8-.6.6-1.8Z" />
    </svg>
  );
}

export function UploadIcon(props: IconProps) {
  return (
    <svg {...baseProps(props)}>
      <path d="M12 15V5" />
      <path d="m7.5 9.5 4.5-4.5 4.5 4.5" />
      <path d="M5 19h14" />
    </svg>
  );
}

export function PlayIcon(props: IconProps) {
  return (
    <svg {...baseProps(props)}>
      <path d="M8 6.5v11l8-5.5-8-5.5Z" />
      <rect x="3.5" y="3.5" width="17" height="17" rx="4.5" />
    </svg>
  );
}

export function HeartIcon(props: IconProps) {
  return (
    <svg {...baseProps(props)}>
      <path d="M12 19s-6.5-3.9-8.4-8c-1.2-2.7.1-6 3.3-6.6 2-.4 3.7.5 5.1 2.3 1.4-1.8 3.1-2.7 5.1-2.3 3.2.6 4.5 3.9 3.3 6.6-1.9 4.1-8.4 8-8.4 8Z" />
    </svg>
  );
}

export function UsersIcon(props: IconProps) {
  return (
    <svg {...baseProps(props)}>
      <circle cx="9" cy="9" r="3" />
      <path d="M4.5 18a4.5 4.5 0 0 1 9 0" />
      <circle cx="17" cy="8" r="2.4" />
      <path d="M14.8 18a3.8 3.8 0 0 1 4.7-3.6 3.7 3.7 0 0 1 2 1.2" />
    </svg>
  );
}

export function DownloadIcon(props: IconProps) {
  return (
    <svg {...baseProps(props)}>
      <path d="M12 5v10" />
      <path d="m7.5 11.5 4.5 4.5 4.5-4.5" />
      <path d="M5 19h14" />
    </svg>
  );
}
