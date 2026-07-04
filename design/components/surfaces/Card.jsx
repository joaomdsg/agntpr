import React from 'react';

/**
 * Bounded surfaces. Everything in Packets sits inside a 1px hairline —
 * nothing floats, nothing drops a shadow (depth is reserved for full
 * app frames and tooltips).
 */
export function Card({ variant = 'card', accent, padding, onClick, style, children }) {
  const V = {
    card: { background: 'var(--surface-card)', border: '1px solid var(--border-faint)', borderRadius: 'var(--r-card)', padding: padding ?? '14px 15px' },
    row: { background: 'var(--surface-panel)', border: '1px solid var(--hairline)', borderRadius: 'var(--r-card)', padding: padding ?? '10px 14px' },
    dashed: { background: 'var(--ground)', border: '1px dashed var(--border-faint)', borderRadius: 'var(--r-card)', padding: padding ?? '14px 15px' },
    tile: { background: 'var(--surface-raised)', border: '1px solid var(--hairline)', borderRadius: 'var(--r-ann)', padding: padding ?? '9px 10px' },
  }[variant];

  return (
    <div
      onClick={onClick}
      style={{
        ...V,
        ...(accent ? { borderLeft: '3px solid ' + accent } : null),
        ...(onClick ? { cursor: 'pointer' } : null),
        ...style,
      }}
    >
      {children}
    </div>
  );
}
