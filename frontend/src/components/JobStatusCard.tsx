interface JobStatusCardProps {
  title: string;
  description: string;
}

export function JobStatusCard({ title, description }: JobStatusCardProps) {
  return (
    <article className="card">
      <h3>{title}</h3>
      <p>{description}</p>
    </article>
  );
}
