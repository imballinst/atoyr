import type { ReactNode } from 'react';

interface StatItem {
  label: ReactNode;
  value: ReactNode;
  valueClassName?: string;
}

export function StatsBar({ stats }: { stats: StatItem[] }) {
  return (
    <div className="bg-dark-bg-tertiary rounded-lg flex divide-x divide-dark-border-primary w-full">
      {stats.map(({ label, value, valueClassName }, index) => (
        <section key={index} className="flex-1 py-2 px-2 text-center">
          <h3 className="text-xs text-dark-text-tertiary mb-1 font-medium">{label}</h3>
          <div className={`text-xl sm:text-3xl font-bold tabular-nums ${valueClassName ?? 'text-dark-interactive-success'}`}>{value}</div>
        </section>
      ))}
    </div>
  );
}
