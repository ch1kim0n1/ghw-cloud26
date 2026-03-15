interface StatItem {
  value: string;
  label: string;
}

interface StatGridProps {
  items: StatItem[];
  className?: string;
}

export function StatGrid({ items, className }: StatGridProps) {
  return (
    <div className={className ?? "stat-grid"}>
      {items.map((item) => (
        <div key={`${item.value}-${item.label}`} className="stat-grid__item">
          <strong>{item.value}</strong>
          <span>{item.label}</span>
        </div>
      ))}
    </div>
  );
}
