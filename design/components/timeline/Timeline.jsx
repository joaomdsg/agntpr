import React from 'react';

const TYPE_COLORS = {
  plan: 'var(--text-muted)',
  edit: 'var(--signal)',
  test: 'var(--held)',
  mutation: 'var(--agent)',
  comment: 'var(--signal)',
  flag: 'var(--agent)',
  held: 'var(--risk)',
  delivered: 'var(--delivered)',
  verified: 'var(--verified)',
};

/**
 * The packet's whole life as a horizontal event track — every action,
 * every author, replayable. Event types map to the state grammar.
 */
export function Timeline({ events = [], selected, onSelect, style }) {
  return (
    <div style={{ display: 'flex', alignItems: 'stretch', position: 'relative', padding: '2px 6px 0', ...style }}>
      <div style={{ position: 'absolute', left: 6, right: 6, top: 24, height: 2, background: 'var(--hairline)' }} />
      {events.map((e, i) => {
        const col = TYPE_COLORS[e.type] || 'var(--text-muted)';
        const big = e.big || e.type === 'held';
        const sel = selected != null && selected === (e.id ?? i);
        const El = onSelect ? 'button' : 'div';
        return (
          <El
            key={e.id ?? i}
            onClick={onSelect ? () => onSelect(e.id ?? i) : undefined}
            style={{
              flex: 1, position: 'relative', zIndex: 1, display: 'flex', flexDirection: 'column',
              alignItems: 'center', gap: 4, paddingTop: 2, background: 'transparent', border: 'none',
              cursor: onSelect ? 'pointer' : 'default', fontFamily: 'var(--font-mono)',
            }}
          >
            <span style={{ fontSize: 9, color: big ? col : 'var(--text-faint)', fontWeight: big || sel ? 700 : 400 }}>{e.t}</span>
            <span
              style={{
                width: big ? 14 : 11, height: big ? 14 : 11, borderRadius: '50%', background: col,
                border: '2px solid var(--surface-panel)',
                boxShadow: big ? `0 0 14px color-mix(in srgb, ${col} 55%, transparent)` : sel ? `0 0 0 2px color-mix(in srgb, ${col} 45%, transparent)` : 'none',
              }}
            />
            <span style={{ fontSize: 9, lineHeight: 1.25, textAlign: 'center', color: big || sel ? 'var(--ink)' : 'var(--text-muted)', fontWeight: big || sel ? 700 : 400, maxWidth: 96 }}>{e.label}</span>
            {e.rev && <span style={{ fontSize: 9, color: 'var(--text-faint)' }}>{e.rev}</span>}
          </El>
        );
      })}
    </div>
  );
}
