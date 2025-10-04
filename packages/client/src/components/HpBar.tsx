import React from 'react';

type Props = {
  value?: number | null;
  max?: number | null;
  fromClass?: string; // e.g. 'from-green-500'
  toClass?: string; // e.g. 'to-lime-400'
  className?: string;
  'aria-label'?: string;
};

export default function HpBar({
  value = 0,
  max = 0,
  fromClass = 'from-green-500',
  toClass = 'to-lime-400',
  className = '',
  'aria-label': ariaLabel,
}: Props) {
  const v = Number(value) || 0;
  const m = Number(max) || 0;
  const pct = m > 0 ? Math.max(0, Math.round((v / m) * 100)) : 0;

  // build gradient class safely
  const gradient = `bg-gradient-to-r ${fromClass} ${toClass}`;

  return (
    <div className={`bg-gray-800 rounded-md p-1 ${className}`} aria-label={ariaLabel}>
      <div className={`h-6 rounded-md transition-all ${gradient}`} style={{ width: `${pct}%` }} />
    </div>
  );
}
