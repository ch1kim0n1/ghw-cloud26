interface SectionHeadingProps {
  eyebrow?: string;
  title: string;
  body?: string;
  compact?: boolean;
}

export function SectionHeading({ eyebrow, title, body, compact = false }: SectionHeadingProps) {
  return (
    <div className={`section-heading${compact ? " section-heading--compact" : ""}`}>
      {eyebrow ? <p className="eyebrow">{eyebrow}</p> : null}
      <h2>{title}</h2>
      {body ? <p>{body}</p> : null}
    </div>
  );
}
