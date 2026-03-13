interface SlotCardProps {
  title: string;
  description: string;
}

export function SlotCard({ title, description }: SlotCardProps) {
  return (
    <article className="card">
      <h3>{title}</h3>
      <p>{description}</p>
    </article>
  );
}
