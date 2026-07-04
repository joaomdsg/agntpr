import React from 'react';

const STATE_COLORS = {
  signal: 'var(--signal)',
  you: 'var(--signal)',
  verified: 'var(--verified)',
  held: 'var(--held)',
  risk: 'var(--risk)',
  agent: 'var(--agent)',
  delivered: 'var(--delivered)',
};

/**
 * Pill-shaped mono microtype chip. State chips are tinted via the
 * color-mix recipe (13–16% fill, optional 40% border); neutral chips
 * are flat --surface-raised (addrs, revs, counts).
 */
export function Chip({ state = 'neutral', size = 'md', bordered = false, dot = false, active = false, uppercase, onClick, title, style, children }) {
  const h = size === 'sm' ? 15 : size === 'lg' ? 20 : 17;
  const px = size === 'sm' ? 7 : size === 'lg' ? 10 : 8;
  const fs = size === 'lg' ? 9.5 : 9;
  const col = STATE_COLORS[state];
  const upper = uppercase != null ? uppercase : state !== 'neutral';

  const base = {
    display: 'inline-flex',
    alignItems: 'center',
    gap: 4,
    height: h,
    padding: `0 ${px}px`,
    borderRadius: h / 2 + 2,
    fontFamily: 'var(--font-mono)',
    fontSize: fs,
    fontWeight: state === 'neutral' ? 600 : 700,
    letterSpacing: upper ? 'var(--track-chip)' : 0,
    textTransform: upper ? 'uppercase' : 'none',
    whiteSpace: 'nowrap',
    cursor: onClick ? 'pointer' : 'default',
    border: '1px solid transparent',
  };

  const look = col
    ? {
        background: `color-mix(in srgb, ${col} 16%, var(--surface-card))`,
        color: col,
        borderColor: bordered ? `color-mix(in srgb, ${col} 40%, var(--surface-card))` : 'transparent',
      }
    : active
      ? { background: 'color-mix(in srgb, var(--signal) 22%, var(--surface-card))', color: 'var(--signal)' }
      : { background: 'var(--surface-raised)', color: 'var(--text-muted)' };

  return (
    <span onClick={onClick} title={title} style={{ ...base, ...look, ...style }}>
      {dot && <span style={{ width: 5, height: 5, borderRadius: '50%', background: 'currentColor', flex: 'none' }} />}
      {children}
    </span>
  );
}
