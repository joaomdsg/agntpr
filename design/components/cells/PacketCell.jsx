import React from 'react';

const STATE_COLORS = {
  inflight: 'var(--signal)',
  you: 'var(--signal)',
  verified: 'var(--verified)',
  held: 'var(--held)',
  risk: 'var(--risk)',
  delivered: 'var(--delivered)',
  agent: 'var(--agent)',
};

/**
 * The atomic unit of the brand: one packet, one square.
 * State is color; live things pulse. `composing` renders the ghost
 * (outline) — below 14px it falls back to solid --delivered-mid
 * (silhouette over story; locked small-size rule).
 */
export function PacketCell({ state = 'inflight', size = 8, live = false, round = false, style }) {
  const radius = round ? '50%' : Math.max(1.5, Math.round(size * 0.23));
  const base = { display: 'inline-block', flex: 'none', width: size, height: size, borderRadius: radius };

  if (state === 'composing') {
    if (size < 14) return <span style={{ ...base, background: 'var(--delivered-mid)', ...style }} />;
    const stroke = Math.max(2, Math.round(size * 0.18));
    return (
      <span
        style={{
          ...base,
          background: 'color-mix(in srgb, var(--delivered) 8%, transparent)',
          border: stroke + 'px solid var(--delivered)',
          boxSizing: 'border-box',
          transform: 'scale(1.035)',
          ...style,
        }}
      />
    );
  }

  const color = STATE_COLORS[state] || state;
  const anim = !live
    ? null
    : state === 'risk'
      ? { animation: 'pk-held-pulse 2s ease-in-out infinite' }
      : { animation: 'pk-pulse 1.8s ease-in-out infinite' };
  const glow = live && state !== 'risk' ? { boxShadow: '0 0 8px ' + (STATE_COLORS[state] || state) } : null;
  return <span style={{ ...base, background: color, ...glow, ...anim, ...style }} />;
}
