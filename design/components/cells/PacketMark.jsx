import React from 'react';
import { PacketCell } from './PacketCell.jsx';

/**
 * The locked 2×2 mark. TL/BL signal, TR ghost (composing), BR delivered.
 * Story: starts as outline top-right, lands as fill bottom-right —
 * same edge, same addr, same color.
 */
export function PacketMark({ cell = 8, wordmark = false, wordmarkSize, sub, held = false, style }) {
  const gap = Math.max(1.5, Math.round(cell * 0.27 * 2) / 2);
  const grid = (
    <span style={{ display: 'grid', gridTemplateColumns: `repeat(2, ${cell}px)`, gridAutoRows: cell + 'px', gap, flex: 'none' }}>
      <PacketCell state="inflight" size={cell} />
      <PacketCell state="composing" size={cell} />
      <PacketCell state={held ? 'risk' : 'inflight'} size={cell} live={held} />
      <PacketCell state="delivered" size={cell} />
    </span>
  );

  if (sub) {
    return (
      <span style={{ display: 'inline-flex', alignItems: 'center', gap: Math.round(cell * 1.2), ...style }}>
        {grid}
        <span style={{ display: 'flex', flexDirection: 'column', gap: 4 }}>
          <span style={{ fontFamily: 'var(--font-mono)', fontSize: 12, fontWeight: 700, letterSpacing: '.03em', color: 'var(--text-muted)', lineHeight: '13px' }}>packets</span>
          <span style={{ fontFamily: 'var(--font-mono)', fontSize: 9, fontWeight: 700, letterSpacing: '.16em', color: 'var(--text-faint)', textTransform: 'uppercase', lineHeight: '9.5px' }}>{sub}</span>
        </span>
      </span>
    );
  }

  if (wordmark) {
    return (
      <span style={{ display: 'inline-flex', alignItems: 'center', gap: Math.round(cell * 1.2), ...style }}>
        {grid}
        <span style={{ fontFamily: 'var(--font-mono)', fontSize: wordmarkSize || Math.round(cell * 1.7), fontWeight: 700, letterSpacing: 'var(--track-word)', color: 'var(--ink)', lineHeight: 1 }}>packets</span>
      </span>
    );
  }

  return <span style={{ display: 'inline-flex', ...style }}>{grid}</span>;
}
